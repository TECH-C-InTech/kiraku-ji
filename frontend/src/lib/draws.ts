import { normalizeApiBaseUrl } from "./api";

export type DrawResponse = {
  post_id: string;
  result: string;
  status: string;
};

type ApiErrorResponse = {
  message?: string;
};

/**
 * 検証済みのおみくじをランダムに取得する。
 */
export const fetchRandomDraw = async (): Promise<DrawResponse> => {
  const response = await fetch(`${normalizeApiBaseUrl()}/draws/random`);

  if (!response.ok) {
    let errorMessage = "おみくじの取得に失敗しました";
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

  return (await response.json()) as DrawResponse;
};
