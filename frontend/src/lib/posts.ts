import { API_BASE_URL } from "./api";

export type CreatePostRequest = {
  post_id: string;
  content: string;
};

export type CreatePostResponse = {
  post_id: string;
};

type ApiErrorResponse = {
  message?: string;
};

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

  const response = await fetch(`${API_BASE_URL}/posts`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
    signal: controller.signal,
  });

  window.clearTimeout(timeoutId);

  if (!response.ok) {
    let errorMessage = "投稿に失敗しました";
    try {
      const data = (await response.json()) as ApiErrorResponse;
      if (data.message) {
        errorMessage = data.message;
      }
    } catch {
      // 何も取得できない場合は既定文言で返す。
    }
    throw new Error(errorMessage);
  }

  return (await response.json()) as CreatePostResponse;
};
