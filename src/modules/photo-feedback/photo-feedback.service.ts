import { Inject, Injectable } from '@nestjs/common';
import { DiscordService } from '@buggins/discord/discord.service';
import { ConfigType } from '@nestjs/config';
import photoFeedbackConfig from './photo-feedback.config';
import { TextChannel } from 'discord.js';

@Injectable()
export class PhotoFeedbackService {
  constructor(
    @Inject(photoFeedbackConfig.KEY)
    private readonly config: ConfigType<typeof photoFeedbackConfig>,
    private readonly discordService: DiscordService,
  ) {}

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
