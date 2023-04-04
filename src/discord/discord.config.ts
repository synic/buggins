import { registerAs } from '@nestjs/config';

export default registerAs('discord', () => ({
  token: process.env.DISCORD_BOT_TOKEN ?? '',
  guildId: process.env.DISCORD_GUILD_ID ?? '',
  clientId: process.env.DISCORD_CLIENT_ID ?? '',
}));
