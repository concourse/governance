# `harmonize`

Synchronizes Discord.

Get it?

Ha ha.

## Setting Up

Instructions for infrastructure team if this ever needs to be set up again:

1. Create a Team in Discord.
    * No non-default settings.
1. Create an Application within the Team.
    * No non-default settings.
1. Create a Bot within the Application.
    * Uncheck "Public Bot" - no one else should be using this.
    * Check "Server Members Intent" - this bot needs the ability to manage all
      members of the server.
1. Copy the bot token and set it as a GitHub secret on `concourse/governance`
   called `DISCORD_ADMIN_BOT_TOKEN`.
1. Under "OAuth2", check the `bot` scope, check the `Administrator` permission,
   and copy the URL.
    * Navigate to the URL in your browser and invite the bot to the server.
1. After the bot has joined, drag their role to the highest level under "Roles"
   in the server settings. This is necessary for the bot to be able to reorder
   roles.

At this point the bot should be a member of the server. It appears offline, so
you'll have to manually check the member list.
