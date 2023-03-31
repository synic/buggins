import { Controller, Inject, Logger } from '@nestjs/common';
import { DiscordService } from './discord.service';
import { Command } from '@squareboat/nest-console';
import discordConfig from './discord.config';
import { ConfigType } from '@nestjs/config';

@Controller()
export class DiscordController {
  private logger = new Logger(DiscordController.name);

  constructor(
    @Inject(discordConfig.KEY)
    private readonly config: ConfigType<typeof discordConfig>,
    private readonly discordService: DiscordService,
  ) {}
}
