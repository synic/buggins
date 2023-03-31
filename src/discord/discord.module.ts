import { DISCORD_CLIENT_PROVIDER } from '@ao/discord/constants';
import { ConfigModule } from '@nestjs/config';
import { Global, ValueProvider } from '@nestjs/common';
import { Module } from '@nestjs/common';
import { Client, GatewayIntentBits, Partials } from 'discord.js';
import discordConfig from './discord.config';
import { DiscordService } from './discord.service';

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

const clientProvider: ValueProvider = {
  provide: DISCORD_CLIENT_PROVIDER,
  useValue: client,
};

@Global()
@Module({
  imports: [ConfigModule.forFeature(discordConfig)],
  providers: [DiscordService, clientProvider],
  exports: [DiscordService, clientProvider],
})
export class DiscordModule {}
