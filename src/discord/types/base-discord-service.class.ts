import { DISCORD_CLIENT_PROVIDER } from '../constants';
import { Inject } from '@nestjs/common';
import { Client, Events, Interaction, SlashCommandBuilder } from 'discord.js';
import { CommandData } from './';

export abstract class BaseDiscordService {
  protected readonly commands = new Map<string, CommandData>();

  constructor(
    @Inject(DISCORD_CLIENT_PROVIDER) protected readonly client: Client,
  ) {
    client.on(Events.InteractionCreate, async (interaction: Interaction) => {
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
    });
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
