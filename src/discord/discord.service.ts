import { Inject, Logger, OnModuleInit } from '@nestjs/common';
import { Command, ConsoleIO } from '@squareboat/nest-console';
import { ConfigType } from '@nestjs/config';
import { Injectable } from '@nestjs/common';
import { REST, Routes } from 'discord.js';
import {
  Client,
  Events,
  GatewayIntentBits,
  Guild,
  Interaction,
  Partials,
  PermissionsBitField,
  SlashCommandBuilder,
} from 'discord.js';
import discordConfig from './discord.config';
import { AddCommandOptions, CommandData } from './types';

@Injectable()
export class DiscordService implements OnModuleInit {
  private logger = new Logger(DiscordService.name);
  readonly #commands = new Map<string, CommandData>();

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

  get commands(): Map<string, CommandData> {
    return this.#commands;
  }

  private setupCommandListener() {
    this.client.on(
      Events.InteractionCreate,
      async (interaction: Interaction) => {
        if (!interaction.isChatInputCommand()) return;

        const command = this.#commands.get(interaction.commandName);

        const permissions =
          typeof interaction.member.permissions === 'string'
            ? { has: () => false }
            : interaction.member.permissions;
        if (
          command.requireMod &&
          !permissions.has(PermissionsBitField.Flags.Administrator)
        ) {
          await interaction.reply({
            content: 'You do not have permission to use this command.',
            ephemeral: true,
          });
          return;
        }

        try {
          await command?.execute(interaction);
          if (command?.autoReply) {
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

  addCommand(options: AddCommandOptions): CommandData {
    const data: CommandData = {
      data: new SlashCommandBuilder()
        .setName(options.name)
        .setDescription(options.description),
      ...options,
    };
    this.#commands.set(options.name, data);
    return data;
  }

  @Command('discord:updatecommands', { desc: 'Update discord commands' })
  async updateCommands(cli: ConsoleIO) {
    const rest = new REST({ version: '10' }).setToken(this.config.token);

    const commands = Array.from(this.#commands.values()).map((c) =>
      c.data.toJSON(),
    );

    cli.info(`Started refreshing ${commands.length} application (/) commands.`);

    const data = (await rest.put(
      Routes.applicationGuildCommands(
        this.config.clientId,
        this.config.guildId,
      ),
      { body: commands },
    )) as { length: number };

    cli.info(`Successfully reloaded ${data.length} application (/) commands.`);
  }
}
