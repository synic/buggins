import { Client, GatewayIntentBits, Partials } from 'discord.js';
import guildMemberAdd from './listeners/guild-member-add';
import ready from './listeners/ready';
import gallery from './listeners/gallery';

const client = new Client({
  intents: [
    GatewayIntentBits.GuildMembers,
    GatewayIntentBits.Guilds,
    GatewayIntentBits.GuildMessages,
    GatewayIntentBits.MessageContent,
    GatewayIntentBits.GuildMessageReactions,
  ],
  partials: [Partials.Message, Partials.Channel, Partials.Reaction],
});

ready(client);
guildMemberAdd(client);
gallery(client);

client.login(process.env.BOT_TOKEN);

console.log(client);
