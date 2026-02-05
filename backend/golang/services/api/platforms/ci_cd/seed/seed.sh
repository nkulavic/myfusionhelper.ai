#!/bin/bash
# Seed CRM platform definitions into DynamoDB
# Usage: ./seed.sh [stage] [region] [profile]
# Example: ./seed.sh dev us-west-2 myfusionhelper

STAGE="${1:-dev}"
REGION="${2:-us-west-2}"
PROFILE="${3:-}"
TABLE_NAME="mfh-${STAGE}-platforms"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Seeding platforms into table: ${TABLE_NAME} (region: ${REGION})"

AWS_CMD="aws dynamodb put-item --region ${REGION}"
if [ -n "${PROFILE}" ]; then
  AWS_CMD="${AWS_CMD} --profile ${PROFILE}"
fi

# Seed each platform
for platform_dir in "${SCRIPT_DIR}"/*/; do
  platform_file="${platform_dir}platform.json"
  if [ ! -f "${platform_file}" ]; then
    continue
  fi

  platform_name=$(jq -r '.name' "${platform_file}")
  platform_id=$(jq -r '.platform_id' "${platform_file}")

  echo "  Seeding: ${platform_name} (${platform_id})..."

  # Convert JSON to DynamoDB format and put item
  ITEM=$(python3 -c "
import json, sys

def to_dynamo(val):
    if isinstance(val, str):
        return {'S': val}
    elif isinstance(val, bool):
        return {'BOOL': val}
    elif isinstance(val, (int, float)):
        return {'N': str(val)}
    elif isinstance(val, list):
        if not val:
            return {'L': []}
        return {'L': [to_dynamo(v) for v in val]}
    elif isinstance(val, dict):
        if not val:
            return {'M': {}}
        return {'M': {k: to_dynamo(v) for k, v in val.items()}}
    elif val is None:
        return {'NULL': True}
    else:
        return {'S': str(val)}

with open('${platform_file}') as f:
    data = json.load(f)

# Add timestamps
from datetime import datetime, timezone
now = datetime.now(timezone.utc).isoformat()
data['created_at'] = now
data['updated_at'] = now

result = {k: to_dynamo(v) for k, v in data.items()}
print(json.dumps(result))
")

  ${AWS_CMD} \
    --table-name "${TABLE_NAME}" \
    --item "${ITEM}" 2>&1

  if [ $? -eq 0 ]; then
    echo "    ✓ ${platform_name} seeded successfully"
  else
    echo "    ✗ Failed to seed ${platform_name}"
  fi
done

echo ""
echo "Seeding complete!"
echo "Verify with: aws dynamodb scan --table-name ${TABLE_NAME} --region ${REGION} ${PROFILE:+--profile ${PROFILE}} --select COUNT"
