import { Inject, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigType } from '@nestjs/config';
import { Injectable } from '@nestjs/common';
import {
  Client,
  Events,
  GatewayIntentBits,
  Guild,
  Interaction,
  Partials,
  SlashCommandBuilder,
} from 'discord.js';
import discordConfig from './discord.config';
import { CommandData } from './types';

@Injectable()
export class DiscordService implements OnModuleInit {
  private logger = new Logger(DiscordService.name);
  protected readonly commands = new Map<string, CommandData>();

  readonly client = new Client({
    intents: [
      GatewayIntentBits.GuildEmojisAndStickers,
      GatewayIntentBits.GuildMembers,
      GatewayIntentBits.GuildMessages,
      GatewayIntentBits.GuildMessageReactions,
      GatewayIntentBits.Guilds,
      GatewayIntentBits.MessageContent,
    ],
    partials: [Partials.Channel, Partials.Message, Partials.Reaction],
  });

  #guild: Guild | undefined | null;

  constructor(
    @Inject(discordConfig.KEY)
    private readonly config: ConfigType<typeof discordConfig>,
  ) {}

  async onModuleInit(): Promise<void> {
    this.client.on(Events.ClientReady, () => {
      this.logger.log(`${this.client.user.username} is online!`);
      this.#guild = this.client.guilds.cache.get(this.config.guildId);
    });

    this.setupCommandListener();
    this.client.login(this.config.token);
  }

  get guild(): Guild | undefined | null {
    return this.#guild;
  }

  private setupCommandListener() {
    this.client.on(
      Events.InteractionCreate,
      async (interaction: Interaction) => {
        if (!interaction.isChatInputCommand()) return;

        const command = this.commands.get(interaction.commandName);

        try {
          await command?.execute(interaction);
          if (command?.autoreply) {
            await interaction.reply('Done!');
          }
        } catch (error) {
          console.error(error);
          if (interaction.replied || interaction.deferred) {
            await interaction.followUp({
              content: 'There was an error while executing this command!',
              ephemeral: true,
            });
          } else {
            await interaction.reply({
              content: 'There was an error while executing this command!',
              ephemeral: true,
            });
          }
        }
      },
    );
  }

  addCommand({
    name,
    description,
    execute,
    autoreply,
  }: {
    name: string;
    description: string;
    execute: (interaction: Interaction) => Promise<void>;
    autoreply?: boolean;
  }) {
    this.commands.set(name, {
      data: new SlashCommandBuilder().setName(name).setDescription(description),
      execute,
      autoreply,
    });
  }
}
