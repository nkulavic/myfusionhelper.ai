# Siri Shortcuts Guide for MyFusion Helper

**Date**: 2026-02-09
**Version**: 1.0
**Platform**: iOS 14+, iPadOS 14+, macOS 11+

---

## Overview

Siri Shortcuts allow you to interact with your MyFusion Helper data using voice commands on Apple devices. Unlike Alexa or Google Assistant, Siri Shortcuts use direct API calls to the chat service—no special webhook needed!

**What you'll need**:
- iPhone, iPad, or Mac with Shortcuts app
- MyFusion Helper account
- API key from your account settings

---

## Setup Instructions

### Step 1: Get Your API Key

1. Log in to [app.myfusionhelper.ai](https://app.myfusionhelper.ai)
2. Navigate to **Settings → API Keys**
3. Click **"Create API Key"**
4. Give it a name: `Siri Shortcuts`
5. Copy the API key (starts with `mfh_`)
6. **Important**: Save this key securely—you won't see it again!

### Step 2: Create Your First Conversation

Before using shortcuts, create a conversation to store your chat history:

1. Open the Shortcuts app on your device
2. Tap **"+"** to create a new shortcut
3. Add **"Get Contents of URL"** action
4. Configure:
   - **URL**: `https://api.myfusionhelper.ai/chat/conversations`
   - **Method**: POST
   - **Headers**:
     - Key: `x-api-key`, Value: `mfh_your_key_here`
     - Key: `Content-Type`, Value: `application/json`
   - **Request Body**: JSON
   ```json
   {"title": "Siri Conversation"}
   ```
5. Add **"Get Dictionary Value"** action
   - **Key**: `conversation_id`
6. Add **"Set Variable"** action
   - **Name**: `ConversationID`
7. Run this shortcut once to create your conversation
8. Note the conversation ID (e.g., `conv:abc-123-def`)

### Step 3: Create the "Ask Fusion Helper" Shortcut

Now create the main shortcut for queries:

1. Create a new shortcut named **"Ask Fusion Helper"**
2. Add **"Ask for Input"** action
   - **Prompt**: `What would you like to know about your CRM data?`
   - **Input Type**: Text
3. Add **"Set Variable"** action
   - **Variable**: `UserQuery`
4. Add **"Text"** action with your conversation ID:
   ```
   conv:YOUR_CONVERSATION_ID_HERE
   ```
5. Add **"Set Variable"** action
   - **Variable**: `ConversationID`
6. Add **"Get Contents of URL"** action
   - **URL**: `https://api.myfusionhelper.ai/chat/conversations/[ConversationID]/messages`
   - **Method**: POST
   - **Headers**:
     - Key: `x-api-key`, Value: `mfh_your_key_here`
     - Key: `Content-Type`, Value: `application/json`
   - **Request Body**: JSON
   ```json
   {"content": "[UserQuery]"}
   ```
7. Add **"Get Text from Input"** action
8. Add **"Show Result"** action (displays response as text)
9. Add **"Speak Text"** action (optional - reads response aloud)

### Step 4: Add Siri Phrase

1. Tap the (i) icon on your shortcut
2. Tap **"Add to Siri"**
3. Record your phrase:
   - **"Ask Fusion Helper"**
   - **"Query my CRM"**
   - **"Check my contacts"**

---

## Example Shortcuts

### Quick Data Query

**Purpose**: Quickly ask about your CRM data

**Actions**:
1. Ask for Input: "What do you want to know?"
2. Get Contents of URL (POST with query)
3. Speak Text (read response)

**Siri Phrase**: "Ask Fusion Helper"

---

### Check Contact Count

**Purpose**: Get contact count without typing

**Actions**:
1. Text: "How many contacts do I have?"
2. Get Contents of URL (POST with fixed query)
3. Show Result

**Siri Phrase**: "How many contacts"

---

### Run Helper

**Purpose**: Execute a helper via voice

**Actions**:
1. Ask for Input: "Which helper do you want to run?"
2. Text: "Run [input] helper"
3. Get Contents of URL (POST)
4. Speak Text

**Siri Phrase**: "Run a helper"

---

### Daily Summary

**Purpose**: Get daily CRM summary

**Actions**:
1. Text: "Give me a summary of my CRM activity"
2. Get Contents of URL (POST)
3. Show Result
4. Speak Text

**Siri Phrase**: "Daily CRM summary"

---

## Advanced: Parsing Streaming Responses

The chat API returns Server-Sent Events (SSE) for streaming. For simple shortcuts, the final aggregated response is sufficient. For advanced users wanting to parse SSE:

1. Use **"Get Contents of URL"** with streaming disabled
2. Add **"Split Text"** action (split by `\n`)
3. Add **"Repeat with Each"** loop
4. Filter for lines starting with `data: `
5. Extract JSON from each line
6. Combine content chunks

*Note: This is complex—most users should use the simple approach.*

---

## Troubleshooting

### "Invalid API key" Error
- Double-check your API key in Settings
- Ensure you're using `x-api-key` header (not `Authorization`)
- API key should start with `mfh_`

### "Conversation not found" Error
- Run the conversation creation shortcut again
- Update the conversation ID in your shortcuts
- Check that you're using the correct conversation ID format

### "No response" or Empty Result
- Check your internet connection
- Verify the API endpoint URL is correct
- Ensure request body JSON is properly formatted

### Response is Not Spoken
- Add **"Speak Text"** action after **"Get Text from Input"**
- Check that Siri voice output is enabled in iOS Settings

---

## API Endpoints Reference

### Create Conversation
```
POST /chat/conversations
Headers: x-api-key: mfh_...
Body: {"title": "My Conversation"}
Response: {"conversation_id": "conv:...", ...}
```

### Send Message
```
POST /chat/conversations/{id}/messages
Headers: x-api-key: mfh_...
Body: {"content": "Your query here"}
Response: SSE stream with content chunks
```

### List Conversations
```
GET /chat/conversations
Headers: x-api-key: mfh_...
Response: {"conversations": [...]}
```

---

## Privacy & Security

- **API keys are sensitive**: Don't share shortcuts containing API keys
- **Conversation history**: All queries are saved in your conversation
- **Data transmission**: All requests use HTTPS encryption
- **Revoke access**: Delete API keys anytime in Settings

---

## Tips & Best Practices

1. **Create separate conversations**: Use different conversations for different purposes (e.g., "Daily Queries", "Helper Execution")
2. **Name your shortcuts descriptively**: Makes them easier to find and use
3. **Test with text first**: Before adding voice, test shortcuts with manual input
4. **Keep queries simple**: Natural language works best (e.g., "show my contacts" vs complex queries)
5. **Use variables**: Store conversation IDs and API keys as variables for reusability

---

## Example Shortcut Downloads

*Coming soon: Download pre-built shortcuts from app.myfusionhelper.ai/shortcuts*

---

## Video Tutorial

*Coming soon: Watch the full video tutorial on our YouTube channel*

---

## Support

If you need help setting up Siri Shortcuts:
- Email: support@myfusionhelper.ai
- Discord: discord.gg/myfusionhelper
- Documentation: docs.myfusionhelper.ai

---

**Last Updated**: 2026-02-09
**Maintained By**: MyFusion Helper Team
