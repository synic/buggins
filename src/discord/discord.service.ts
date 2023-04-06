import { Inject, Injectable, Logger } from '@nestjs/common';
import { Once, InjectDiscordClient } from '@discord-nestjs/core';
import { Channel, Client, Events, Guild } from 'discord.js';
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

  findDiscordGuild(): Guild | null {
    return this.client.guilds.cache.get(this.config.guildId) ?? null;
  }

  findDiscordChannelByName<T extends Channel>(name: string): T | null {
    return this.findDiscordGuild()?.channels.cache.find(
      (c) => c.name.toLowerCase() === name.toLowerCase(),
    ) as T;
  }

  @Once(Events.ClientReady)
  async onReady() {
    this.logger.log(`'${this.client.user?.username}' is connected to Discord!`);
  }
}
