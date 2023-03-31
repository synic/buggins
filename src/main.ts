import { Client, Events, GatewayIntentBits, Partials } from 'discord.js';
import { ENABLED_MODULES } from './constants';
import { init } from './storage';
import { DiscordModule } from './types';

const client = new Client({
  intents: [
    GatewayIntentBits.GuildEmojisAndStickers,
    GatewayIntentBits.GuildMembers,
    GatewayIntentBits.GuildMessages,
    GatewayIntentBits.GuildMessageReactions,
    GatewayIntentBits.Guilds,
    GatewayIntentBits.MessageContent,
  ],
  partials: [Partials.Channel, Partials.Message, Partials.Reaction],
});

ENABLED_MODULES.forEach(
  (e: { new (client: Client): DiscordModule }) => new e(client),
);

client.on(Events.ClientReady, async () => {
  await init();
  if (!client.user || !client.application) {
    return;
  }

  console.log(`${client.user.username} is online`);
});

client.login(process.env.BOT_TOKEN);
