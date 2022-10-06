import { Client, GuildMember } from 'discord.js';

export default (client: Client): void => {
  client.on('guildMemberAdd', async (member: GuildMember) => {
    console.log(`${member.nickname} has joined`);
  });
};
