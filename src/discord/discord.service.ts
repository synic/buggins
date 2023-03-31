import { Inject, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigType } from '@nestjs/config';
import { Injectable } from '@nestjs/common';
import { Client, Events } from 'discord.js';
import { DISCORD_CLIENT_PROVIDER } from './constants';
import discordConfig from './discord.config';

@Injectable()
export class DiscordService implements OnModuleInit {
  private logger = new Logger(DiscordService.name);

  constructor(
    @Inject(discordConfig.KEY)
    private readonly config: ConfigType<typeof discordConfig>,
    @Inject(DISCORD_CLIENT_PROVIDER) private readonly client: Client,
  ) {}

  async onModuleInit(): Promise<void> {
    this.client.on(Events.ClientReady, () => {
      this.logger.log(`${this.client.user.username} is online!`);
    });
    this.client.login(this.config.token);
  }
}
