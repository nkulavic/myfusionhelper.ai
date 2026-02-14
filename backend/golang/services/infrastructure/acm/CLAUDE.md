# ACM Certificate Infrastructure

ACM (AWS Certificate Manager) service for managing SSL/TLS certificates for custom domains.

## Overview

This service manages ACM certificates for the MyFusion Helper API Gateway custom domain. It uses DNS validation via Route53 for automatic certificate verification.

## Certificate Configuration

**Primary Domain**: `api.myfusionhelper.ai` (production)
**Stage Domains**: `api-dev.myfusionhelper.ai`, `api-staging.myfusionhelper.ai`

The certificate includes both the primary domain and stage-specific subdomains as Subject Alternative Names (SANs), allowing a single certificate to cover multiple environments.

## DNS Validation

The certificate uses DNS validation method with automatic validation via Route53:

```yaml
ValidationMethod: DNS
DomainValidationOptions:
  - DomainName: api.myfusionhelper.ai
    HostedZoneId: Z071462818IPQJBH38AMK  # myfusionhelper.ai hosted zone
```

CloudFormation automatically creates the required CNAME records in Route53 for validation. The certificate typically validates within 5-10 minutes of deployment.

## CloudFormation Exports

This service exports the following values for use by other services:

- `CertificateArn`: ARN of the ACM certificate (used by API Gateway)
- `CertificateDomainName`: Primary domain name (`api.myfusionhelper.ai`)
- `CertificateStageDomainName`: Stage-specific domain name

## Deployment

```bash
cd backend/golang/services/infrastructure/acm
npx sls deploy --stage dev
```

**Note**: This service must deploy **before** the API Gateway service, as the gateway references the certificate ARN via CloudFormation exports.

## Certificate Lifecycle

- **Renewal**: ACM automatically renews certificates before expiration (60 days before)
- **Validation**: DNS validation records persist in Route53, allowing automatic renewal
- **Stages**: Each stage (dev, staging, main) has its own certificate stack

## Verification

Check certificate status:

```bash
# Get certificate ARN from CloudFormation exports
aws cloudformation describe-stacks \
  --stack-name mfh-infrastructure-acm-dev \
  --query "Stacks[0].Outputs[?OutputKey=='CertificateArn'].OutputValue" \
  --output text

# Check certificate details
aws acm describe-certificate \
  --certificate-arn <arn> \
  --region us-west-2
```

## Dependencies

**Requires**:
- Route53 hosted zone for `myfusionhelper.ai` (ID: Z071462818IPQJBH38AMK)

**Used by**:
- API Gateway service (references certificate ARN)

## Deployment Order

1. Deploy ACM service (creates certificate)
2. Wait for validation (~5-10 minutes)
3. Deploy API Gateway service (uses certificate for custom domain)
4. Deploy Route53 service (creates DNS records pointing to gateway)
