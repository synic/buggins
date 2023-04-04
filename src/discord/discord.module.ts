import { ConfigModule, ConfigType } from '@nestjs/config';
import { Module } from '@nestjs/common';
import discordConfig from './discord.config';
import { GatewayIntentBits, Partials } from 'discord.js';
import { DiscordModule as BaseDiscordModule } from '@discord-nestjs/core';
import { DiscordService } from './discord.service';

@Module({
  imports: [
    ConfigModule.forFeature(discordConfig),
    BaseDiscordModule.forRootAsync({
      imports: [ConfigModule.forFeature(discordConfig)],
      inject: [discordConfig.KEY],
      useFactory: (config: ConfigType<typeof discordConfig>) => ({
        token: config.token,
        discordClientOptions: {
          intents: [
            GatewayIntentBits.GuildEmojisAndStickers,
            GatewayIntentBits.GuildMembers,
            GatewayIntentBits.GuildMessages,
            GatewayIntentBits.GuildMessageReactions,
            GatewayIntentBits.Guilds,
            GatewayIntentBits.MessageContent,
          ],
          partials: [Partials.Channel, Partials.Message, Partials.Reaction],
        },
        registerCommandOptions: [
          {
            forGuild: config.guildId,
          },
        ],
      }),
    }),
  ],
  providers: [DiscordService],
  exports: [DiscordService, BaseDiscordModule.forFeature()],
})
export class DiscordModule {}
