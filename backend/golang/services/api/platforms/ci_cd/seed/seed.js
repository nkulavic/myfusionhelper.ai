#!/usr/bin/env node
/**
 * Seed platform definitions into DynamoDB
 * Usage: STAGE=dev node seed.js
 * Example: STAGE=prod AWS_REGION=us-west-2 node seed.js
 */

const { DynamoDBClient } = require('@aws-sdk/client-dynamodb')
const { DynamoDBDocumentClient, PutCommand } = require('@aws-sdk/lib-dynamodb')
const fs = require('fs')
const path = require('path')

// Configuration from environment variables (GitHub Actions friendly)
const STAGE = process.env.STAGE || 'dev'
const REGION = process.env.AWS_REGION || 'us-west-2'
const TABLE_NAME = `mfh-${STAGE}-platforms`

console.log(`🌱 Platform Seeding Script`)
console.log(`   Stage: ${STAGE}`)
console.log(`   Platforms Table: ${TABLE_NAME}`)
console.log(`   Region: ${REGION}`)
console.log(``)

// Initialize DynamoDB client
const client = new DynamoDBClient({ region: REGION })
const docClient = DynamoDBDocumentClient.from(client)

async function seedPlatform(platformDir) {
  const platformFile = path.join(platformDir, 'platform.json')

  if (!fs.existsSync(platformFile)) {
    return null
  }

  const platformData = JSON.parse(fs.readFileSync(platformFile, 'utf8'))
  const platformName = platformData.name
  const platformId = platformData.platform_id

  console.log(`  Seeding: ${platformName} (${platformId})...`)

  // Add timestamps
  const now = new Date().toISOString()
  platformData.created_at = now
  platformData.updated_at = now

  try {
    await docClient.send(
      new PutCommand({
        TableName: TABLE_NAME,
        Item: platformData,
      })
    )
    console.log(`    ✓ ${platformName} seeded successfully`)
    return platformName
  } catch (error) {
    console.error(`    ✗ Failed to seed ${platformName}:`, error.message)
    return null
  }
}

async function main() {
  const seedDir = __dirname
  const entries = fs.readdirSync(seedDir, { withFileTypes: true })

  const platformDirs = entries
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(seedDir, entry.name))

  const results = await Promise.all(platformDirs.map(seedPlatform))
  const successful = results.filter(Boolean).length
  const total = platformDirs.length

  console.log('')
  console.log('Seeding complete!')
  console.log(`  Seeded ${successful}/${total} platforms successfully`)
  console.log(
    `  Verify with: aws dynamodb scan --table-name ${TABLE_NAME} --region ${REGION} --select COUNT`
  )

  if (successful < total) {
    process.exit(1)
  }
}

main().catch((error) => {
  console.error('Fatal error:', error)
  process.exit(1)
})
