import { Inject, Injectable } from '@nestjs/common';
import {
  MessageReaction,
  Message,
  Events,
  User,
  TextChannel,
} from 'discord.js';
import { On } from '@discord-nestjs/core';
import { DiscordService } from '@ao/discord/discord.service';
import galleryConfig from './gallery.config';
import { ConfigType } from '@nestjs/config';

@Injectable()
export class GalleryService {
  constructor(
    @Inject(galleryConfig.KEY)
    private readonly config: ConfigType<typeof galleryConfig>,
    private readonly discordService: DiscordService,
  ) {}

  @On(Events.MessageCreate)
  async onMessageCreate(message: Message) {
    const [galleryChannel, _] = this.getChannels();

    if (message.channelId === galleryChannel?.id && !message.author.bot) {
      if (message.attachments.size === 1) {
        await message.react('👍');
        await message.react('👎');
        await message.react('💬');
      } else {
        await message.delete();
      }
    }
  }

  @On(Events.MessageReactionAdd)
  async onMessageReactionAdd(reaction: MessageReaction, user: User) {
    const [galleryChannel, feedbackChannel] = this.getChannels();

    const attachment = reaction.message.attachments.first();
    if (
      !user.bot &&
      reaction.emoji.name === '💬' &&
      reaction.message.channelId === galleryChannel?.id &&
      reaction.message.author &&
      attachment
    ) {
      await (feedbackChannel as TextChannel).send({
        content:
          `Hey ${reaction.message.author.toString()}, ${user.toString()} would like to give you ` +
          `feedback on your image.\n Click here for the original: ${reaction.message.url}`,
        files: [attachment],
      });
    }
  }

  getChannels(): [TextChannel | null, TextChannel | null] {
    return [
      this.discordService.findChannelByName<TextChannel>(
        this.config.postChannelName,
      ),
      this.discordService.findChannelByName<TextChannel>(
        this.config.feedbackChannelName,
      ),
    ];
  }
}
