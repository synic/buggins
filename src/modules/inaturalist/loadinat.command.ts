import { Injectable } from '@nestjs/common';
import { Command, Handler } from '@discord-nestjs/core';
import { CommandInteraction, PermissionsBitField } from 'discord.js';
import { INaturalistService } from './inaturalist.service';

@Command({
  name: 'loadinat',
  description: 'Load a random inaturalist observation.',
})
@Injectable()
export class LoadInatCommand {
  constructor(private readonly inaturalistService: INaturalistService) {}

  @Handler()
  async handler(interaction: CommandInteraction): Promise<string> {
    const permissions =
      typeof interaction.member?.permissions === 'string'
        ? { has: () => false }
        : interaction.member?.permissions ?? { has: () => false };

    if (!permissions.has(PermissionsBitField.Flags.Administrator)) {
      return 'Only administrators can use this command';
    }

    await this.inaturalistService.fetchAll();
    return 'Done!';
  }
}
