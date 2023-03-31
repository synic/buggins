import { DiscordModule } from '@ao/discord/discord.module';
import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import photoFeedbackConfig from './photo-feedback.config';
import { PhotoFeedbackService } from './photo-feedback.service';

@Module({
  imports: [ConfigModule.forFeature(photoFeedbackConfig), DiscordModule],
  providers: [PhotoFeedbackService],
  exports: [PhotoFeedbackService],
})
export class PhotoFeedbackModule {}
