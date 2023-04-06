import { Module } from '@nestjs/common';
import { ConsoleModule } from '@squareboat/nest-console';
import { DatabasesModule } from './databases/databases.module';
import { DiscordModule } from './discord/discord.module';
import { ScheduleModule } from '@nestjs/schedule';
import { INaturalistModule } from './modules/inaturalist/inaturalist.module';

@Module({
  imports: [
    ConsoleModule,
    DatabasesModule,
    DiscordModule,
    ScheduleModule.forRoot(),

    // bot modules
    INaturalistModule,
  ],
})
export class AppModule {}
