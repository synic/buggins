import { DiscordModule } from '@buggins/discord/discord.module';
import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import photoFeedbackConfig from './photo-feedback.config';
import { PhotoFeedbackListener } from './photo-feedback.listener';
import { PhotoFeedbackService } from './photo-feedback.service';

@Module({
  imports: [ConfigModule.forFeature(photoFeedbackConfig), DiscordModule],
  providers: [PhotoFeedbackService, PhotoFeedbackListener],
  exports: [PhotoFeedbackService],
})
export class PhotoFeedbackModule {}
