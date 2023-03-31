import { DiscordModule } from '@ao/discord/discord.module';
import { Module } from '@nestjs/common';
import { GalleryService } from './gallery.service';

@Module({
  imports: [DiscordModule],
  providers: [GalleryService],
  exports: [GalleryService],
})
export class GalleryModule {}
