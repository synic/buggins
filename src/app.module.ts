import { Module } from '@nestjs/common';
import { ConsoleModule } from '@squareboat/nest-console';
import { DatabasesModule } from './databases/databases.module';
import { DiscordModule } from './discord/discord.module';
import { INaturalistModule } from './modules/inaturalist/inaturalist.module';
import { ScheduleModule } from '@nestjs/schedule';

const enabledModules = [INaturalistModule];

@Module({
  imports: [
    ConsoleModule,
    DatabasesModule,
    DiscordModule,
    ScheduleModule.forRoot(),

    ...enabledModules,
  ],
})
export class AppModule {}
