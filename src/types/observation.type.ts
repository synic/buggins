export type Observation = {
  id: number;
  photos: {
    url: string;
    user: {
      login: string;
    };
  }[];
};
