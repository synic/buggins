import { Inject, Injectable } from '@nestjs/common';
import {
  MessageReaction,
  Message,
  Events,
  User,
  TextChannel,
} from 'discord.js';
import { On } from '@discord-nestjs/core';
import { DiscordService } from '@buggins/discord/discord.service';
import { ConfigType } from '@nestjs/config';
import photoFeedbackConfig from './photo-feedback.config';

@Injectable()
export class PhotoFeedbackService {
  constructor(
    @Inject(photoFeedbackConfig.KEY)
    private readonly config: ConfigType<typeof photoFeedbackConfig>,
    private readonly discordService: DiscordService,
  ) {}

  @On(Events.MessageCreate)
  async onMessageCreate(message: Message) {
    const [postChannel, _] = this.getChannels();

    if (message.channelId === postChannel?.id && !message.author.bot) {
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
    const [postChannel, feedbackChannel] = this.getChannels();

    const attachment = reaction.message.attachments.first();
    if (
      !user.bot &&
      reaction.emoji.name === '💬' &&
      reaction.message.channelId === postChannel?.id &&
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
