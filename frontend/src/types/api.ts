export type ApiErrorResponse = {
  message?: string;
};

export type CreatePostRequest = {
  post_id: string;
  content: string;
};

export type CreatePostResponse = {
  post_id: string;
};

export type DrawResponse = {
  post_id: string;
  result: string;
  status: string;
};
