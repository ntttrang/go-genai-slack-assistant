# Slack Bot Setup Guide
## Required Bot Token Scopes

To fix this, go to your Slack App settings (https://api.slack.com/apps) and add these **Bot Token Scopes**:

### Required Scopes:
1. **`channels:history`** - Read messages in public channels
2. **`channels:read`** - View basic information about public channels  
3. **`chat:write`** - Post messages in channels
4. **`groups:history`** - Read messages in private channels (if needed)
5. **`groups:read`** - View basic information about private channels (if needed)
6. **`im:history`** - Read messages in direct messages (if needed)
7. **`users:read`** - View people in a workspace
8. **`reactions:write`** - Required to add emoji reactions
9. **`reactions:read`** - Optional, to read reaction data

### Steps to Add Scopes:

1. Go to https://api.slack.com/apps
2. Select your app "my-assistant"
3. Go to **"OAuth & Permissions"** in the left sidebar
4. Scroll down to **"Scopes"** section
5. Under **"Bot Token Scopes"**, click **"Add an OAuth Scope"**
6. Add all the scopes listed above
7. Scroll to top and click **"Reinstall to Workspace"**
8. Approve the permissions
9. Copy the new **Bot User OAuth Token** (starts with `xoxb-`)
10. Update your `.env` file with the new token:
    ```bash
    SLACK_BOT_TOKEN=xoxb-your-new-token-here
    ```
11. Setup Webhook endpoint on Slack:
   Event Subscriptions > Toggle "Enable Events" to ON > Add Request URL 
      ``` bash
      https://xxxx-xxx-xxx.ngrok.io/slack/events
      ```
      *** If it says "Verified" ✅, you're good
      *** If it fails ❌, your server might not be accessible or not running

*** If your server start on local, use ngrok to public host ( for testing only)
    ```bash
       ngrok http 8080
    ```
#### Then use the ngrok URL in Slack: https://xxxx-xxx-xxx.ngrok.io/slack/events


### After Updating Scopes:

1. Restart your bot:
   ```bash
   pkill -f slack-bot
   source .env
   nohup ./bin/slack-bot > server.log 2>&1 &
   ```

2. Make sure the bot is in the channel:
   - In Slack, go to #chatchit channel
   - Type `/invite @my-assistant`
   - The bot should join the channel
3. Start server
4. Test again:
   - Execute bash script
   ```bash
   ./test.sh
   ```
   - Or type message at Slack channel