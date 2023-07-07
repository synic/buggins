import { DiscordService } from '@buggins/discord/discord.service';
import { On, Once } from '@discord-nestjs/core';
import { Inject, Injectable } from '@nestjs/common';
import { ConfigType } from '@nestjs/config';
import { Message, TextChannel } from 'discord.js';
import thisthatConfig from './thisthat.config';

const reactMap: { [key: string]: string } = {
  '1': 'one',
  '2': 'two',
  '3': 'three',
  '4': 'four',
  '5': 'five',
  '6': 'six',
  '7': 'seven',
  '8': 'eight',
  '9': 'nine',
  '10': 'ten',
};

@Injectable()
export class ThisThatService {
  private channel: TextChannel | null;

  constructor(
    @Inject(thisthatConfig.KEY)
    private readonly config: ConfigType<typeof thisthatConfig>,
    private readonly discordService: DiscordService,
  ) {}

  @Once('ready')
  async onReady(): Promise<void> {
    this.channel = this.discordService.findDiscordChannelByName(
      this.config.channelName,
    );
  }

  @On('messageCreate')
  async onMessage(message: Message): Promise<void> {
    if (
      message.channelId === this.channel?.id &&
      message.attachments.size > 1
    ) {
      for (let i = 1; i < message.attachments.size + 1; i++) {
        await message.react(`\:${reactMap[i.toString()]}:`);
      }
    }
  }
}
