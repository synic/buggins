import { DiscordModule } from '@buggins/discord/discord.module';
import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import thisthatConfig from './thisthat.config';

@Module({
  imports: [ConfigModule.forFeature(thisthatConfig), DiscordModule],
})
export class ThisThatModule {}
