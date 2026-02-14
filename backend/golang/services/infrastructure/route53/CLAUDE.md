# Route53 DNS Infrastructure

Route53 service for managing DNS records for MyFusion Helper custom domains.

## Overview

This service manages DNS records that point custom domains to the API Gateway regional endpoint. It creates alias records that route traffic from user-friendly domains like `api.myfusionhelper.ai` to the API Gateway.

## DNS Configuration

**Hosted Zone**: `myfusionhelper.ai` (ID: Z071462818IPQJBH38AMK)

**Records Created**:
- **Production**: `api.myfusionhelper.ai` → API Gateway (when stage=prod)
- **Dev**: `api-dev.myfusionhelper.ai` → API Gateway (when stage=dev)
- **Staging**: `api-staging.myfusionhelper.ai` → API Gateway (when stage=staging)

## Alias Records

The service creates Route53 alias records that point to the API Gateway regional domain name:

```yaml
Type: A
AliasTarget:
  HostedZoneId: ${cf:mfh-api-gateway-dev.RegionalHostedZoneId}  # API Gateway's hosted zone
  DNSName: ${cf:mfh-api-gateway-dev.RegionalDomainName}        # API Gateway's domain
  EvaluateTargetHealth: false
```

Alias records are preferred over CNAME records because:
- No additional DNS lookup costs
- Can be used at the zone apex
- Automatic health checking (if enabled)
- Better performance

## Stage-Specific Domains

The service uses CloudFormation conditions to create different records per stage:

- **Production** (`prod`): Creates `api.myfusionhelper.ai`
- **Non-production** (`dev`, `staging`): Creates `api-{stage}.myfusionhelper.ai`

This allows each environment to have its own subdomain while production uses the clean `api.` subdomain.

## CloudFormation Exports

This service exports:

- `ApiDomainName`: The custom domain name for this stage
- `HostedZoneId`: The Route53 hosted zone ID

## Deployment

```bash
cd backend/golang/services/infrastructure/route53
npx sls deploy --stage dev
```

**Note**: This service must deploy **after** the API Gateway service, as it references CloudFormation exports from the gateway (regional domain name and hosted zone ID).

## Verification

Check DNS propagation:

```bash
# Query DNS directly
dig api-dev.myfusionhelper.ai

# Use nslookup
nslookup api-dev.myfusionhelper.ai

# Check Route53 records
aws route53 list-resource-record-sets \
  --hosted-zone-id Z071462818IPQJBH38AMK \
  --query "ResourceRecordSets[?Name=='api-dev.myfusionhelper.ai.']"
```

Test the API:

```bash
# Health check
curl -v https://api-dev.myfusionhelper.ai/health

# Auth health check
curl -v https://api-dev.myfusionhelper.ai/auth/health
```

## DNS Propagation

After deployment:
- DNS changes propagate within 60 seconds (TTL: 60)
- Alias records resolve immediately within AWS
- External DNS resolvers may cache up to 5 minutes

## Dependencies

**Requires**:
- Route53 hosted zone for `myfusionhelper.ai`
- API Gateway service (provides regional domain name via CloudFormation exports)

**Used by**:
- Frontend applications (use custom domain for API calls)
- External integrations
- Documentation

## Deployment Order

1. Deploy infrastructure services (including ACM)
2. Deploy API Gateway service (creates custom domain)
3. **Deploy Route53 service** (creates DNS record)
4. Wait for DNS propagation (~60 seconds)
5. Test endpoints via custom domain

## Troubleshooting

**DNS not resolving:**
- Check CloudFormation stack status: `aws cloudformation describe-stacks --stack-name mfh-infrastructure-route53-dev`
- Verify API Gateway exports exist: `aws cloudformation list-exports | grep mfh-api-gateway`
- Check Route53 record: `aws route53 list-resource-record-sets --hosted-zone-id Z071462818IPQJBH38AMK`

**502 Bad Gateway:**
- Verify API Gateway custom domain is healthy
- Check API Gateway deployment status
- Verify certificate is valid and attached

**SSL/TLS errors:**
- Verify ACM certificate is issued and validated
- Check certificate covers the domain (SANs)
- Ensure using HTTPS (not HTTP)
