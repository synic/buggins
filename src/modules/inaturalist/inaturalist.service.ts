import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigType } from '@nestjs/config';
import { CronJob } from 'cron';
import { SchedulerRegistry } from '@nestjs/schedule';
import { In, Repository } from 'typeorm';
import { InjectRepository } from '@nestjs/typeorm';
import { EmbedBuilder, Guild, TextChannel } from 'discord.js';
import { Observation } from './types';
import { Result, Ok } from 'ts-results';
import { FetchCommunicationError } from '@buggins/common/types';
import { httpRequest, shuffleArray } from '@buggins/common/utils';
import { SeenObservationEntity } from './seen-observation.entity';
import inaturalistConfig from './inaturalist.config';
import { DiscordService } from '@buggins/discord/discord.service';
import { INaturalistSettingsEntity } from './inaturalist-settings.entity';

@Injectable()
export class INaturalistService implements OnModuleInit {
  private readonly logger = new Logger(INaturalistService.name);
  private readonly runningJobs: CronJob[] = [];

  constructor(
    private readonly discordService: DiscordService,
    @Inject(inaturalistConfig.KEY)
    private readonly config: ConfigType<typeof inaturalistConfig>,
    @InjectRepository(SeenObservationEntity)
    private readonly seenObservationsRepository: Repository<SeenObservationEntity>,
    @InjectRepository(INaturalistSettingsEntity)
    private readonly inaturalistSettingsRepository: Repository<INaturalistSettingsEntity>,
    private readonly schedulerRegistry: SchedulerRegistry,
  ) {}

  async onModuleInit(): Promise<void> {
    await this.refresh();
  }

  async stopAllJobs(): Promise<void> {
    while (this.runningJobs.length > 0) {
      const job = this.runningJobs.pop();
      job?.stop();
    }
  }

  async refresh(): Promise<void> {
    await this.stopAllJobs();
    const enabledSettings = await this.findEnabledGuildSettings();

    for (const settings of enabledSettings) {
      if (settings.cronPattern) {
        this.logger.log(
          `Setting up iNaturalist cron schedule for '${settings.guildEntity?.name}' ` +
            `with schedule '${settings.cronPattern}'.`,
        );

        const job = new CronJob(settings.cronPattern, () =>
          this.fetch(settings.guildId),
        );
        this.schedulerRegistry.addCronJob(
          `inaturalist-fetch-${settings.guildEntity?.id}`,
          job,
        );
        job.start();
        this.runningJobs.push(job);
      }
    }
  }

  async findEnabledGuildSettings(): Promise<INaturalistSettingsEntity[]> {
    return await this.inaturalistSettingsRepository.findBy({ isEnabled: true });
  }

  private async fetchRecentProjectObservations(
    projectId: string,
  ): Promise<Result<Observation[], FetchCommunicationError>> {
    const response = await httpRequest<Observation[]>({
      server: 'https://inaturalist.org',
      path: `observations/project/${projectId}.json?order_by=id&order=desc&per_page=50`,
    });

    if (!response.ok) return response;

    return Ok(response.val);
  }

  private async markObservationAsSeen(
    guildId: string,
    observation: Observation,
  ): Promise<void> {
    await this.seenObservationsRepository.save({
      guildId,
      observationId: observation.id.toString(),
    });
  }

  private async showObservation(
    channel: TextChannel,
    observation: Observation,
  ): Promise<void> {
    const photoUrl = observation.photos[0].large_url;
    const embed = new EmbedBuilder({
      description: `**[${observation.user_login}](https://inaturalist.org/people/${observation.user_id}) has spotted something new!**`,
    });

    embed.addFields([
      {
        name: `Taxon`,
        value: `${
          (observation.taxon?.common_name?.name ??
            observation.taxon?.default_name?.name ??
            observation.species_guess) ||
          'unknown'
        }`,
      },
      {
        name: 'iNaturalist Link',
        value: `https://inaturalist.org/observations/${observation.id}`,
      },
      {
        name: 'iNaturalist Project',
        value: `https://inaturalist.org/projects/${this.config.projectId}`,
      },
    ]);
    embed.setImage(photoUrl);

    await channel?.send({ embeds: [embed] });
    await this.markObservationAsSeen(channel.guild.id, observation);

    return;
  }

  async findSettingsForGuildId(
    guildId: string,
  ): Promise<INaturalistSettingsEntity | null> {
    return await this.inaturalistSettingsRepository.findOneBy({
      guildId,
    });
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
      (o) => !seenObservationIds.includes(o.id.toString()),
    );

    this.logger.log(`Seen observation count is ${seenObservationIds.length}`);
    this.logger.log(`Unseen observation count is ${unseen.length}`);

    if (unseen.length <= 0) {
      this.logger.log(`No unseen observations left to post.`);
      return null;
    }

    const userObservationMap = new Map<number, Observation[]>();

    for (const observation of unseen) {
      let userObservations: Observation[] | null =
        userObservationMap.get(observation.user_id) ?? null;
      if (userObservations == null) {
        userObservations = [];
        userObservationMap.set(observation.user_id, userObservations);
      }

      userObservations.push(observation);
    }

    const user = shuffleArray(Array.from(userObservationMap.keys()))[0];
    const userArray = userObservationMap.get(user);

    if (userArray == null) {
      this.logger.warn(`User array was null`);
      return null;
    }

    this.logger.log(`User is ${user}, items for user is ${userArray.length}`);

    return shuffleArray(userArray)[0];
  }

  async fetchAll(): Promise<void> {
    const enabledSettings = await this.findEnabledGuildSettings();
    enabledSettings.forEach(async (s) => await this.fetch(s.guildId));
  }

  async fetch(guildOrGuildId: string | Guild): Promise<void> {
    const guild =
      typeof guildOrGuildId === 'string'
        ? this.discordService.findDiscordGuild(guildOrGuildId)
        : guildOrGuildId;

    if (!guild) {
      this.logger.warn(
        `Unable to find guild for fetching observations for '${guildOrGuildId}.`,
      );
      return;
    }

    const settings = await this.findSettingsForGuildId(guild.id);

    if (!settings) {
      this.logger.warn(`Unable to find settings for guild id '${guild.id}'.`);
      return;
    }

    const observationsResponse = await this.fetchRecentProjectObservations(
      settings.projectId,
    );

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

    const channel = this.discordService.findChannelByName<TextChannel>(
      settings.guildId,
      settings.channelName,
    );

    if (!channel) {
      this.logger.warn(
        `Unable to find channel for settings id '${settings.id}'.`,
      );
      return;
    }

    await this.showObservation(channel, observation);
  }
}
