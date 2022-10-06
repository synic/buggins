import { Client, GatewayIntentBits } from 'discord.js';
import guildMemberAdd from './listeners/guild-member-add';
import ready from './listeners/ready';

const client = new Client({
  intents: [
    GatewayIntentBits.GuildMembers,
    GatewayIntentBits.Guilds,
    GatewayIntentBits.GuildMessages,
    GatewayIntentBits.MessageContent,
  ],
});

ready(client);
guildMemberAdd(client);

client.login(process.env.BOT_TOKEN);

console.log(client);
