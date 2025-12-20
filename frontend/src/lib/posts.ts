import { getApiBaseUrl, getApiErrorMessage } from "./api";

export type CreatePostRequest = {
  post_id: string;
  content: string;
};

export type CreatePostResponse = {
  post_id: string;
};

// APIのベースURLの末尾のスラッシュを取り除く。
const normalizeApiBaseUrl = () => getApiBaseUrl().replace(/\/+$/, "");

/**
 * 闇投稿を登録する。
 */
export const createPost = async (content: string) => {
  const postId = crypto.randomUUID();
  const payload: CreatePostRequest = {
    post_id: postId,
    content,
  };

  const controller = new AbortController();
  const timeoutId = window.setTimeout(() => controller.abort(), 10_000);

  const response = await fetch(`${normalizeApiBaseUrl()}/posts`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
    signal: controller.signal,
  });

  window.clearTimeout(timeoutId);

  if (!response.ok) {
    const errorMessage = await getApiErrorMessage(
      response,
      "投稿に失敗しました",
    );
    throw new Error(errorMessage);
  }

  return (await response.json()) as CreatePostResponse;
};
