import { ConfigModule, ConfigType } from '@nestjs/config';
import { Module } from '@nestjs/common';
import discordConfig from './discord.config';
import { GatewayIntentBits } from 'discord.js';
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
            GatewayIntentBits.GuildMessages,
            GatewayIntentBits.Guilds,
            GatewayIntentBits.MessageContent,
          ],
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
