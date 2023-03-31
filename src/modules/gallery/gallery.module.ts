import { Module } from '@nestjs/common';
import { GalleryService } from './gallery.service';

@Module({
  providers: [GalleryService],
  exports: [GalleryService],
})
export class GalleryModule {}
