import { API_BASE_URL, getApiErrorMessage, getNetworkErrorMessage } from "./api";

export type DrawResponse = {
  post_id: string;
  result: string;
  status: string;
};

/**
 * 検証済みのおみくじをランダムに取得する。
 */
export const fetchRandomDraw = async (): Promise<DrawResponse> => {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}/draws/random`);
  } catch (error) {
    throw new Error(getNetworkErrorMessage(error));
  }

  if (!response.ok) {
    const errorMessage = await getApiErrorMessage(
      response,
      "おみくじの取得に失敗しました",
    );
    throw new Error(errorMessage);
  }

  return (await response.json()) as DrawResponse;
};
