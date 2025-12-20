import type { DrawResponse } from "@/types/api";
import { getApiErrorMessageFromResponse } from "@/utils/api";
import { getApiBaseUrl, normalizeApiBaseUrl } from "./api";

/**
 * 検証済みのおみくじをランダムに取得する。
 */
export const fetchRandomDraw = async (): Promise<DrawResponse> => {
  const response = await fetch(`${getApiBaseUrl()}/draws/random`);

  if (!response.ok) {
    const errorMessage = await getApiErrorMessageFromResponse(
      response,
      "おみくじの取得に失敗しました",
    );
    throw new Error(errorMessage);
  }

  return (await response.json()) as DrawResponse;
};
