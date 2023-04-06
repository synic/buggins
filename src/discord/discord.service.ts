import { Injectable, Logger } from '@nestjs/common';
import { Once, InjectDiscordClient } from '@discord-nestjs/core';
import { Channel, Client, Events, Guild } from 'discord.js';
import { GuildEntity } from './guild.entity';

@Injectable()
export class DiscordService {
  private readonly logger = new Logger(DiscordService.name);

  constructor(@InjectDiscordClient() private readonly client: Client) {}

  findDiscordGuild(guildIdOrGuildentity: string | GuildEntity): Guild | null {
    const id =
      typeof guildIdOrGuildentity === 'string'
        ? guildIdOrGuildentity
        : guildIdOrGuildentity.id;
    return this.client.guilds.cache.get(id) ?? null;
  }

  findChannelByName<T extends Channel>(
    guildOrGuildId: string | Guild,
    name: string,
  ): T | null {
    const guild =
      typeof guildOrGuildId === 'string'
        ? this.findDiscordGuild(guildOrGuildId)
        : guildOrGuildId;
    return guild?.channels.cache.find(
      (c) => c.name.toLowerCase() === name.toLowerCase(),
    ) as T;
  }

  @Once(Events.ClientReady)
  async onReady() {
    this.logger.log(`'${this.client.user?.username}' is connected to Discord!`);
  }
}
