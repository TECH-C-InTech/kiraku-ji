import { API_BASE_URL, getApiErrorMessage } from "./api";

export type DrawResponse = {
  post_id: string;
  result: string;
  status: string;
};

/**
 * 検証済みのおみくじをランダムに取得する。
 */
export const fetchRandomDraw = async (): Promise<DrawResponse> => {
  const response = await fetch(`${API_BASE_URL}/draws/random`);

  if (!response.ok) {
    const errorMessage = await getApiErrorMessage(
      response,
      "おみくじの取得に失敗しました",
    );
    throw new Error(errorMessage);
  }

  return (await response.json()) as DrawResponse;
};
