import { Module } from '@nestjs/common';
import { ConsoleModule } from '@squareboat/nest-console';
import { DatabasesModule } from './databases/databases.module';
import { DiscordModule } from './discord/discord.module';
import { GalleryModule } from './modules/gallery/gallery.module';
import { INaturalistModule } from './modules/inaturalist/inaturalist.module';

const enabledModules = [INaturalistModule, GalleryModule];

@Module({
  imports: [ConsoleModule, DatabasesModule, DiscordModule, ...enabledModules],
})
export class AppModule {}
