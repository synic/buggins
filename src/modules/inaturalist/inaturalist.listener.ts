import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { SchedulerRegistry } from '@nestjs/schedule';
import { CronJob } from 'cron';
import { ConfigType } from '@nestjs/config';
import inaturalistConfig from './inaturalist.config';
import { INaturalistService } from './inaturalist.service';

@Injectable()
export class INaturalistListener implements OnModuleInit {
  private readonly logger = new Logger(INaturalistService.name);

  constructor(
    @Inject(inaturalistConfig.KEY)
    private readonly config: ConfigType<typeof inaturalistConfig>,
    private readonly inaturalistService: INaturalistService,
    private readonly schedulerRegistry: SchedulerRegistry,
  ) {}

  onModuleInit(): void {
    const job = new CronJob(this.config.cronPattern, () =>
      this.inaturalistService.fetch(),
    );
    this.schedulerRegistry.addCronJob('inaturalist-fetch', job);
    job.start();

    this.logger.log(
      `Set up iNaturalist fetch cronjob with pattern: ${this.config.cronPattern}`,
    );
  }
}
