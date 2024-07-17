import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { CronJob } from 'cron';
import { ConfigType } from '@nestjs/config';
import { In, Repository } from 'typeorm';
import { InjectRepository } from '@nestjs/typeorm';
import { EmbedBuilder, TextChannel } from 'discord.js';
import { Observation } from './types';
import { Result, Ok } from 'ts-results';
import { FetchCommunicationError } from '@buggins/common/types';
import { httpRequest, shuffleArray } from '@buggins/common/utils';
import { SeenObservationEntity } from './seen-observation.entity';
import inaturalistConfig from './inaturalist.config';
import { DiscordService } from '@buggins/discord/discord.service';
import { SchedulerRegistry } from '@nestjs/schedule';

@Injectable()
export class INaturalistService implements OnModuleInit {
  private readonly logger = new Logger(INaturalistService.name);
  private readonly pageSize = 200;
  private readonly displayedObservers = new Set<number>();

  constructor(
    private readonly discordService: DiscordService,
    @Inject(inaturalistConfig.KEY)
    private readonly config: ConfigType<typeof inaturalistConfig>,
    @InjectRepository(SeenObservationEntity)
    private readonly seenObservationsRepository: Repository<SeenObservationEntity>,
    private readonly schedulerRegistry: SchedulerRegistry,
  ) {}

  onModuleInit(): void {
    const job = new CronJob(this.config.cronPattern, () => this.fetch());
    this.schedulerRegistry.addCronJob('inaturalist-fetch', job);
    job.start();

    this.logger.log(
      `Set up fetch cronjob with pattern: ${this.config.cronPattern}`,
    );
  }

  private async fetchRecentProjectObservations(): Promise<
    Result<Observation[], FetchCommunicationError>
  > {
    const observations: Observation[] = [];

    const maxPages = 10;
    let tries = 0;

    while (tries < maxPages) {
      const response = await httpRequest<Observation[]>({
        server: 'https://inaturalist.org',
        path: `observations/project/${this.config.projectId}.json?order_by=id&order=desc&per_page=${this.pageSize}`,
      });

      if (!response.ok) {
        if (observations.length > 0) return Ok(observations);
        return response;
      }

      response.val
        .filter((o) => {
          return (
            (o.photos || []).length > 0 &&
            ![undefined, ''].includes(o.photos[0].large_url)
          );
        })
        .forEach((o) => observations.push(o));

      if (response.val.length < this.pageSize) {
        return Ok(observations);
      }
      tries++;
    }
    return Ok(observations);
  }

  private async markObservationAsSeen(o: Observation): Promise<void> {
    await this.seenObservationsRepository.save({ observationId: o.id });
    if (!this.displayedObservers.has(o.user_id)) {
      this.displayedObservers.add(o.user_id);
    }

    this.logger.log(`displayedObservers ${this.displayedObservers}`);
  }

  private async showObservation(o: Observation): Promise<void> {
    const channel = this.discordService.findDiscordChannelByName<TextChannel>(
      this.config.channelName,
    );

    if (!channel) {
      this.logger.error(
        `Could not find channel named: ${this.config.channelName}`,
      );
      return;
    }

    const photoUrl = o.photos[0].large_url;
    const embed = new EmbedBuilder({
      description: `**[${o.user_login}](https://inaturalist.org/people/${o.user_id}) has spotted something new!**`,
    });

    embed.addFields([
      {
        name: `Taxon`,
        value: `${o?.taxon?.name ?? 'unknown'} (${
          (o.taxon?.common_name?.name ??
            o.taxon?.default_name?.name ??
            o.species_guess) ||
          'unknown'
        })`,
      },
      {
        name: 'iNaturalist Link',
        value: `https://inaturalist.org/observations/${o.id}`,
      },
      {
        name: 'Our community iNaturalist Project',
        value: `https://inaturalist.org/projects/${this.config.projectId}`,
      },
    ]);
    embed.setImage(photoUrl);

    await channel?.send({ embeds: [embed] });
    await this.markObservationAsSeen(o);

    return;
  }

  async selectRandomObservation(
    observations: Observation[],
  ): Promise<Observation | null> {
    const observationIds = observations.map((o) => o.id);
    const seenObservationIds = (
      await this.seenObservationsRepository.findBy({
        observationId: In(observationIds),
      })
    ).map((o) => o.observationId);

    const unseen = observations.filter(
      (o) => !seenObservationIds.includes(o.id),
    );

    this.logger.log(`Seen observation count is ${seenObservationIds.length}`);
    this.logger.log(`Unseen observation count is ${unseen.length}`);

    if (unseen.length <= 0) {
      this.logger.log(`No unseen observations left to post.`);
      return null;
    }

    const observerObservationMap = new Map<number, Observation[]>();

    for (const observation of unseen) {
      let userObservations: Observation[] | null =
        observerObservationMap.get(observation.user_id) ?? null;
      if (userObservations == null) {
        userObservations = [];
        observerObservationMap.set(observation.user_id, userObservations);
      }

      userObservations.push(observation);
    }

    const observers = new Set(observerObservationMap.keys());
    let potentialObservers = new Set(
      [...observers].filter((o) => !this.displayedObservers.has(o)),
    );

    if (potentialObservers.size <= 0) {
      potentialObservers = new Set(observers);
      this.displayedObservers.clear();
    }

    const observer = shuffleArray(Array.from(observerObservationMap.keys()))[0];
    const observerArray = observerObservationMap.get(observer);

    if (observerArray == null) {
      this.logger.warn(`User array was null`);
      return null;
    }

    this.logger.log(
      `User is ${observer}, items for user is ${observerArray.length}`,
    );

    return shuffleArray(observerArray)[0];
  }

  async fetch(): Promise<void> {
    const observationsResponse = await this.fetchRecentProjectObservations();

    if (!observationsResponse.ok) {
      this.logger.error(
        `Error fetching observations: ${observationsResponse.val}`,
      );
      return;
    }

    const observation = await this.selectRandomObservation(
      observationsResponse.val,
    );

    if (!observation) {
      this.logger.log(`No unseen observations to display at this time.`);
      return;
    }
    await this.showObservation(observation);
  }
}
