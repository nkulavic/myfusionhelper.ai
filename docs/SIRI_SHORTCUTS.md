# Siri Shortcuts Setup Guide

Use Siri to query your CRM data using MyFusion Helper's REST API. No app download required - just the built-in iOS/macOS Shortcuts app.

## Prerequisites

- iPhone, iPad, or Mac with iOS 16+ / macOS 13+
- MyFusion Helper account with active subscription
- API key (generated from your dashboard)

## Quick Start

1. **Generate an API Key**
   - Log in to [MyFusion Helper Dashboard](https://app.myfusionhelper.ai)
   - Navigate to Settings → API Keys
   - Click "Create API Key"
   - Give it a name (e.g., "Siri Shortcuts")
   - Copy the key (starts with `mfh_...`)
   - **Important**: Save this key securely - it won't be shown again

2. **Create Your First Shortcut**
   - Open the Shortcuts app on your device
   - Tap "+" to create a new shortcut
   - Follow the configuration steps below

---

## Basic Query Shortcut

This shortcut queries your CRM data using natural language.

### Configuration Steps

1. **Add "Get Contents of URL" Action**
   - Search for "Get Contents of URL" in the actions list
   - Tap to add it to your shortcut

2. **Configure the API Request**
   - **URL**: `https://api.myfusionhelper.ai/chat/conversations`
   - **Method**: `POST`
   - **Headers**:
     ```
     x-api-key: mfh_your_api_key_here
     Content-Type: application/json
     ```
   - **Request Body**:
     ```json
     {
       "title": "Siri Query",
       "metadata": {
         "source": "siri_shortcut"
       }
     }
     ```

3. **Parse Conversation ID**
   - Add "Get Dictionary Value" action
   - Key: `data.conversation_id`
   - Input: Contents of URL (from previous step)

4. **Send Your Query**
   - Add another "Get Contents of URL" action
   - **URL**: `https://api.myfusionhelper.ai/chat/conversations/[Dictionary Value]/messages`
   - **Method**: `POST`
   - **Headers**:
     ```
     x-api-key: mfh_your_api_key_here
     Content-Type: application/json
     ```
   - **Request Body**:
     ```json
     {
       "content": "Show my Keap contacts"
     }
     ```

5. **Extract Response**
   - Add "Get Dictionary Value" action
   - Key: `data.content`
   - Input: Contents of URL (from step 4)

6. **Show Result**
   - Add "Show Result" action
   - Input: Dictionary Value (from step 5)

7. **Name Your Shortcut**
   - Tap the shortcut name at the top
   - Enter: "Check CRM Contacts"

8. **Add to Siri**
   - Tap the (i) info button
   - Tap "Add to Siri"
   - Record your phrase: "Check my contacts"

---

## Advanced: Dynamic Query Shortcut

This shortcut asks you what to query each time you run it.

### Configuration Steps

1. **Ask for Input**
   - Add "Ask for Input" action
   - Prompt: "What would you like to know about your CRM?"
   - Input Type: Text
   - Default Answer: (leave blank)

2. **Create Conversation** (same as steps 1-3 above)
   - Add "Get Contents of URL" action
   - URL: `https://api.myfusionhelper.ai/chat/conversations`
   - Method: POST
   - Headers and body same as basic shortcut

3. **Parse Conversation ID**
   - Add "Get Dictionary Value" action
   - Key: `data.conversation_id`

4. **Send Dynamic Query**
   - Add "Get Contents of URL" action
   - URL: `https://api.myfusionhelper.ai/chat/conversations/[Dictionary Value]/messages`
   - Method: POST
   - Headers: Same as above
   - **Request Body**:
     ```json
     {
       "content": "[Provided Input]"
     }
     ```
   - Tap `[Provided Input]` and select the variable from "Ask for Input"

5. **Extract and Show Response** (same as steps 5-6 above)

6. **Add to Siri**
   - Name: "Ask Fusion Helper"
   - Siri phrase: "Ask my CRM assistant"

---

## Helper Invocation Shortcut

Run MyFusion Helper automation helpers via Siri.

### Example: Tag a Contact

1. **Ask for Contact ID**
   - Add "Ask for Input" action
   - Prompt: "What's the contact ID?"
   - Input Type: Number

2. **Create Conversation and Send Helper Command**
   - Follow steps 1-3 from Basic Query Shortcut
   - In step 4, use this request body:
     ```json
     {
       "content": "Tag contact [Provided Input] as VIP in Keap"
     }
     ```

3. **Show Confirmation**
   - Extract response content
   - Add "Show Result" or "Speak Text" action

---

## Voice-Only Shortcut (Hands-Free)

Perfect for in-car use or when your screen isn't accessible.

### Configuration

1. **Follow Dynamic Query Shortcut steps 1-4**

2. **Replace "Show Result" with "Speak Text"**
   - Add "Speak Text" action instead of "Show Result"
   - Input: Dictionary Value (the response content)
   - This will read the response aloud

3. **Enable Voice Confirmation**
   - Add "Set Siri Response" action (optional)
   - Input: "Here's what I found:"

4. **Add to Siri**
   - Name: "Voice CRM Query"
   - Siri phrase: "Check my CRM"

---

## Connection-Specific Queries

Query specific CRM platforms by including the connection name in your query.

### Example Request Bodies

**Query Keap contacts:**
```json
{
  "content": "Show my Keap contacts created in the last 7 days"
}
```

**Query GoHighLevel opportunities:**
```json
{
  "content": "Show my GoHighLevel opportunities with value over $1000"
}
```

**Query ActiveCampaign tags:**
```json
{
  "content": "List all tags in ActiveCampaign"
}
```

---

## API Reference

### Create Conversation

**Endpoint**: `POST https://api.myfusionhelper.ai/chat/conversations`

**Headers**:
```
x-api-key: mfh_your_api_key_here
Content-Type: application/json
```

**Request Body**:
```json
{
  "title": "Conversation Title",
  "metadata": {
    "source": "siri_shortcut",
    "device": "iPhone"
  }
}
```

**Response**:
```json
{
  "success": true,
  "message": "Conversation created",
  "data": {
    "conversation_id": "conv_abc123xyz",
    "title": "Conversation Title",
    "created_at": "2026-02-09T12:00:00Z"
  }
}
```

### Send Message

**Endpoint**: `POST https://api.myfusionhelper.ai/chat/conversations/{conversation_id}/messages`

**Headers**:
```
x-api-key: mfh_your_api_key_here
Content-Type: application/json
```

**Request Body**:
```json
{
  "content": "Your natural language query here"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Message sent",
  "data": {
    "message_id": "msg_xyz789abc",
    "role": "assistant",
    "content": "You have 142 contacts in Keap. The 5 most recent are...",
    "tool_calls": [
      {
        "tool": "query_crm_data",
        "status": "success"
      }
    ],
    "created_at": "2026-02-09T12:00:05Z"
  }
}
```

---

## Troubleshooting

### "Unauthorized" Error

**Problem**: Response returns 401 Unauthorized

**Solutions**:
- Verify your API key is correct (starts with `mfh_`)
- Check the key hasn't been revoked in your dashboard
- Ensure the header is exactly `x-api-key` (lowercase)

### "Invalid Request" Error

**Problem**: Response returns 400 Bad Request

**Solutions**:
- Verify `Content-Type: application/json` header is set
- Check JSON formatting in request body (use a validator)
- Ensure conversation_id is correctly extracted from previous step

### No Response or Timeout

**Problem**: Shortcut hangs or times out

**Solutions**:
- Check your internet connection
- Verify API endpoint URLs are correct (https, not http)
- Try a simpler query first to isolate the issue

### Siri Can't Find Shortcut

**Problem**: Siri says she can't find the shortcut

**Solutions**:
- Open Shortcuts app and verify the shortcut exists
- Re-record your Siri phrase (Settings → Siri & Search → My Shortcuts)
- Use a unique phrase that doesn't conflict with other shortcuts

---

## Example Queries

Here are natural language queries you can use:

### Contact Queries
- "Show my Keap contacts"
- "How many contacts do I have in GoHighLevel?"
- "Show contacts created in the last week"
- "Find contacts tagged as VIP"

### Helper Queries
- "List all my helpers"
- "Run the RFM Calculation helper on contact 12345"
- "Tag contact 67890 as Premium in Keap"

### Data Queries
- "Show my ActiveCampaign email open rates"
- "What are my top performing campaigns?"
- "Show contacts with lifetime value over $5000"

### Summary Queries
- "Give me a summary of my CRM activity today"
- "Show my dashboard metrics"
- "What's my contact count across all platforms?"

---

## Privacy & Security

### API Key Security

- **Never share your API key** with anyone
- **Don't post shortcuts online** that contain your API key
- If compromised, revoke the key immediately in your dashboard
- Use separate API keys for different devices/purposes

### Data Privacy

- Queries are processed by MyFusion Helper servers
- No data is stored on Apple's servers
- Conversation history is retained for 90 days
- Delete conversations anytime from your dashboard

### Shortcut Sharing

To share a shortcut:
1. Remove your API key before exporting
2. Add instructions for users to insert their own key
3. Use placeholder text like `YOUR_API_KEY_HERE`

---

## Tips & Best Practices

### Performance

- **Reuse conversations**: Store conversation_id in a variable for multiple queries
- **Be specific**: More specific queries return faster, more relevant results
- **Limit results**: Add "limit to 10" in your query to reduce response time

### Voice Commands

- **Use clear phrases**: Choose Siri phrases that are distinct and easy to say
- **Avoid conflicts**: Don't use phrases similar to built-in Siri commands
- **Test variations**: Try different phrasings to see what works best

### Organization

- **Group shortcuts**: Create folders in the Shortcuts app (e.g., "CRM Queries")
- **Use descriptive names**: Make it easy to find shortcuts later
- **Add notes**: Use the shortcut description field to document usage

---

## Advanced Use Cases

### Morning CRM Briefing

Create a shortcut that runs automatically every morning:

1. Create shortcut with multiple queries:
   - "Show new contacts from yesterday"
   - "Show upcoming appointments today"
   - "Show open opportunities"
2. Combine responses into a briefing
3. Set as morning automation (Settings → Automation)

### Contact Lookup Widget

Add a home screen widget for quick contact lookups:

1. Create query shortcut
2. Add to home screen
3. Configure widget to show recent results

### Integration with Other Apps

Combine with other shortcuts:
- Send CRM data to Notes app
- Create calendar events from CRM data
- Forward contact info via Messages

---

## Support

Need help? Contact support:
- **Email**: support@myfusionhelper.ai
- **Dashboard**: Settings → Help & Support
- **Documentation**: https://docs.myfusionhelper.ai

---

## Version History

- **1.0.0** (2026-02-09): Initial documentation
  - Basic query shortcut
  - Dynamic query shortcut
  - Helper invocation shortcut
  - Voice-only shortcut
  - API reference
  - Troubleshooting guide
