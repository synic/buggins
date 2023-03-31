import { DiscordModule } from '@ao/discord/discord.module';
import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import galleryConfig from './gallery.config';
import { GalleryService } from './gallery.service';

@Module({
  imports: [ConfigModule.forFeature(galleryConfig), DiscordModule],
  providers: [GalleryService],
  exports: [GalleryService],
})
export class GalleryModule {}
