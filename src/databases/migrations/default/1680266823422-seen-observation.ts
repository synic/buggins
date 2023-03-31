import { MigrationInterface, QueryRunner } from "typeorm";

export class seenObservation1680266823422 implements MigrationInterface {
    name = 'seenObservation1680266823422'

    public async up(queryRunner: QueryRunner): Promise<void> {
        await queryRunner.query(`CREATE TABLE "inaturalist_seen_observation" ("id" integer PRIMARY KEY AUTOINCREMENT NOT NULL, "created_at" datetime NOT NULL DEFAULT (datetime('now')), "updated_at" datetime NOT NULL DEFAULT (datetime('now')), "observation_id" varchar NOT NULL)`);
    }

    public async down(queryRunner: QueryRunner): Promise<void> {
        await queryRunner.query(`DROP TABLE "inaturalist_seen_observation"`);
    }

}
