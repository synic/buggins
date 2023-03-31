export type Observation = {
  id: number;
  user: {
    login: string;
  };
  photos: {
    large_url: string;
  }[];
};
