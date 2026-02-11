# Custom Domain Setup for API Gateway

## Overview

This document describes the custom domain setup for the MyFusion Helper API Gateway, enabling user-friendly URLs like `api.myfusionhelper.ai` instead of the default AWS-generated endpoint.

## Architecture

```
User Request (HTTPS)
    ↓
DNS (Route53): api-dev.myfusionhelper.ai
    ↓
Alias Record → API Gateway Custom Domain
    ↓
API Gateway Regional Endpoint
    ↓
Lambda Functions (Auth, Helpers, Data, etc.)
```

## Components

### 1. ACM Certificate (`services/infrastructure/acm`)

**Purpose**: Manages SSL/TLS certificates for HTTPS connections

**Configuration**:
- Primary domain: `api.myfusionhelper.ai`
- Stage domains: `api-dev.myfusionhelper.ai`, `api-staging.myfusionhelper.ai`
- Validation: DNS (automatic via Route53)
- Renewal: Automatic (ACM handles renewal)

**CloudFormation Exports**:
- `CertificateArn`: Used by API Gateway for custom domain

### 2. API Gateway Custom Domain (`services/api/gateway`)

**Purpose**: Maps custom domain to API Gateway

**Configuration**:
```yaml
httpApi:
  domain:
    domainName: api-dev.myfusionhelper.ai  # (or api.myfusionhelper.ai for prod)
    certificateArn: ${cf:mfh-infrastructure-acm-dev.CertificateArn}
    endpointType: REGIONAL
    securityPolicy: TLS_1_2
```

**CloudFormation Exports**:
- `RegionalDomainName`: API Gateway's regional endpoint (e.g., `d-abc123.execute-api.us-west-2.amazonaws.com`)
- `RegionalHostedZoneId`: API Gateway's hosted zone ID for alias records

### 3. Route53 DNS Records (`services/infrastructure/route53`)

**Purpose**: Routes DNS queries to API Gateway

**Configuration**:
- Hosted Zone: `myfusionhelper.ai` (ID: Z071462818IPQJBH38AMK)
- Record Type: A (Alias)
- Target: API Gateway regional endpoint

**Stage-specific domains**:
- Dev: `api-dev.myfusionhelper.ai`
- Staging: `api-staging.myfusionhelper.ai`
- Production: `api.myfusionhelper.ai`

## Deployment Order

**Critical**: Services must deploy in this exact order:

1. **ACM** (`mfh-infrastructure-acm-{stage}`)
   - Creates certificate
   - Sets up DNS validation records
   - Waits for validation (~5-10 minutes)

2. **API Gateway** (`mfh-api-gateway-{stage}`)
   - References ACM certificate ARN
   - Creates custom domain mapping
   - Exports regional domain name and hosted zone ID

3. **Route53** (`mfh-infrastructure-route53-{stage}`)
   - References API Gateway exports
   - Creates DNS alias record
   - Propagates DNS changes (~60 seconds)

## CI/CD Integration

The deployment workflow (`.github/workflows/deploy-backend.yml`) enforces the correct order:

```yaml
jobs:
  deploy-infra:          # Includes ACM
    matrix:
      - acm
      - cognito
      - dynamodb-core
      # ...

  deploy-gateway:        # Depends on deploy-infra
    needs: deploy-infra

  deploy-route53:        # Depends on deploy-gateway
    needs: deploy-gateway

  deploy-api:            # Depends on deploy-gateway
    needs: deploy-gateway
```

## Manual Deployment

Deploy each service individually (for testing):

```bash
# 1. Deploy ACM certificate
cd backend/golang/services/infrastructure/acm
npx sls deploy --stage dev

# Wait for certificate validation (check AWS Console or CLI)
aws acm describe-certificate --certificate-arn <arn> --region us-west-2

# 2. Deploy API Gateway (once cert is validated)
cd backend/golang/services/api/gateway
npx sls deploy --stage dev

# 3. Deploy Route53 DNS records
cd backend/golang/services/infrastructure/route53
npx sls deploy --stage dev
```

## Verification

### 1. Check Certificate Status

```bash
# Get certificate ARN
aws cloudformation describe-stacks \
  --stack-name mfh-infrastructure-acm-dev \
  --query "Stacks[0].Outputs[?OutputKey=='CertificateArn'].OutputValue" \
  --output text

# Check validation status
aws acm describe-certificate \
  --certificate-arn <arn> \
  --region us-west-2 \
  --query "Certificate.Status"
```

Expected: `"ISSUED"`

### 2. Check API Gateway Custom Domain

```bash
# Get custom domain details
aws apigatewayv2 get-domain-name \
  --domain-name api-dev.myfusionhelper.ai \
  --region us-west-2

# Check API mapping
aws apigatewayv2 get-api-mappings \
  --domain-name api-dev.myfusionhelper.ai \
  --region us-west-2
```

### 3. Check DNS Resolution

```bash
# Query DNS directly
dig api-dev.myfusionhelper.ai

# Expected output:
# api-dev.myfusionhelper.ai. 60 IN A <alias-to-api-gateway>

# Test with nslookup
nslookup api-dev.myfusionhelper.ai
```

### 4. Test HTTPS Endpoints

```bash
# Health check (public endpoint)
curl -v https://api-dev.myfusionhelper.ai/health

# Expected: 200 OK with JSON response

# Auth health check
curl -v https://api-dev.myfusionhelper.ai/auth/health

# Expected: 200 OK
```

## Frontend Integration

Update frontend environment variables to use custom domain:

### Development

```bash
# apps/web/.env.local
NEXT_PUBLIC_API_URL=https://api-dev.myfusionhelper.ai
```

### Production

```bash
# apps/web/.env.production
NEXT_PUBLIC_API_URL=https://api.myfusionhelper.ai
```

## Troubleshooting

### Issue: Certificate stuck in "Pending Validation"

**Cause**: DNS validation records not created or Route53 hosted zone issue

**Solution**:
1. Check Route53 for CNAME validation records
2. Verify hosted zone ID matches in ACM serverless.yml
3. Wait up to 10 minutes for DNS propagation

### Issue: 502 Bad Gateway on custom domain

**Cause**: API Gateway custom domain not mapped correctly

**Solution**:
1. Check API mapping exists: `aws apigatewayv2 get-api-mappings --domain-name <domain>`
2. Verify API Gateway stage is deployed: `aws apigatewayv2 get-stages --api-id <id>`
3. Check CloudFormation stack for errors

### Issue: DNS not resolving

**Cause**: Route53 alias record missing or incorrect

**Solution**:
1. Check Route53 record: `aws route53 list-resource-record-sets --hosted-zone-id Z071462818IPQJBH38AMK`
2. Verify CloudFormation exports exist: `aws cloudformation list-exports | grep mfh-api-gateway`
3. Wait 60 seconds for DNS propagation

### Issue: SSL/TLS certificate errors

**Cause**: Certificate not valid for domain or not attached

**Solution**:
1. Verify certificate covers domain: `aws acm describe-certificate --certificate-arn <arn>`
2. Check SubjectAlternativeNames include your domain
3. Ensure using HTTPS (not HTTP)

## Cost Impact

**ACM Certificate**: Free (AWS-managed certificates)

**Route53**:
- Hosted zone: $0.50/month
- Queries: $0.40 per million queries
- Alias queries to AWS resources: Free

**API Gateway Custom Domain**: No additional cost

**Estimated monthly cost**: ~$0.50 (Route53 hosted zone only)

## Production Considerations

### Multi-Stage Setup

The infrastructure supports multiple stages with different domains:

- **Dev**: `api-dev.myfusionhelper.ai` → `mfh-api-gateway-dev`
- **Staging**: `api-staging.myfusionhelper.ai` → `mfh-api-gateway-staging`
- **Production**: `api.myfusionhelper.ai` → `mfh-api-gateway-prod`

Each stage has its own:
- ACM certificate stack
- API Gateway custom domain
- Route53 DNS record

### Certificate Renewal

ACM automatically renews certificates 60 days before expiration. No manual intervention required, as long as:
- DNS validation records remain in Route53
- Hosted zone is active
- Certificate is in use (attached to API Gateway)

### DNS TTL

The Route53 alias record has no explicit TTL (AWS manages it). Typical resolution time:
- Within AWS: Immediate
- External resolvers: 60 seconds (based on AWS's internal TTL)

## Security

### TLS Version

API Gateway custom domain enforces TLS 1.2 minimum:

```yaml
securityPolicy: TLS_1_2
```

This ensures strong encryption for all HTTPS connections.

### Certificate Validation

DNS validation is more secure than email validation:
- No email interception risk
- Automated renewal works reliably
- Validation records can be version controlled

### Default Endpoint

The default API Gateway endpoint (`https://a95gb181u4.execute-api.us-west-2.amazonaws.com`) remains accessible. To disable it:

```yaml
httpApi:
  disableExecuteApiEndpoint: true  # Uncomment after custom domain is verified
```

**Recommendation**: Keep default endpoint enabled during initial rollout, then disable after custom domain is verified working.

## Rollback Plan

If custom domain setup fails or causes issues:

1. **Frontend**: Revert `NEXT_PUBLIC_API_URL` to default endpoint
2. **Route53**: Delete DNS record via CloudFormation
3. **API Gateway**: Remove custom domain configuration
4. **ACM**: Delete certificate (if not in use)

Default API Gateway endpoint continues working throughout the process.

## Next Steps

After custom domain setup is complete:

1. Update frontend environment variables
2. Update API documentation with new URLs
3. Configure CORS if needed (already configured in gateway)
4. Set up monitoring/alarms for custom domain endpoints
5. Update external integrations (webhooks, OAuth redirects, etc.)

## References

- [AWS ACM Documentation](https://docs.aws.amazon.com/acm/)
- [API Gateway Custom Domains](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html)
- [Route53 Alias Records](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/resource-record-sets-choosing-alias-non-alias.html)
- [Serverless Framework HTTP API](https://www.serverless.com/framework/docs/providers/aws/events/http-api)
