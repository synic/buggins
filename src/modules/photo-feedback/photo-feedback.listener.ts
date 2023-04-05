import { Injectable } from '@nestjs/common';
import { Message, Events } from 'discord.js';
import { On } from '@discord-nestjs/core';
import { PhotoFeedbackService } from './photo-feedback.service';

@Injectable()
export class PhotoFeedbackListener {
  constructor(private readonly photoFeedbackService: PhotoFeedbackService) {}

  @On(Events.MessageCreate)
  async onMessageCreate(message: Message) {
    const [postChannel, _] = this.photoFeedbackService.getChannels();

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
}
