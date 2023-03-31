import { Inject, Injectable, Logger } from '@nestjs/common';
import { Once, InjectDiscordClient } from '@discord-nestjs/core';
import { Client, Events, Guild } from 'discord.js';
import discordConfig from './discord.config';
import { ConfigType } from '@nestjs/config';

@Injectable()
export class DiscordService {
  private readonly logger = new Logger(DiscordService.name);

  constructor(
    @Inject(discordConfig.KEY)
    private readonly config: ConfigType<typeof discordConfig>,
    @InjectDiscordClient() private readonly client: Client,
  ) {}

  getGuild(): Guild | null {
    return this.client.guilds.cache.get(this.config.guildId) ?? null;
  }

  @Once(Events.ClientReady)
  async onReady() {
    this.logger.log(`Bot ${this.client.user.tag} was started!`);
  }
}
