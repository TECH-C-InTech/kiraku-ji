import type { ApiErrorResponse } from "@/types/api";

const statusMessageMap: Record<number, string> = {
  400: "入力が正しくありません",
  404: "データが見つかりません",
  409: "すでに登録されています",
  500: "サーバーで問題が発生しました",
};

/**
 * API レスポンスのメッセージを取得し、見つからなければ既定文言を返す。
 */
export const getApiErrorMessageFromResponse = async (
  response: Response,
  fallbackMessage: string,
): Promise<string> => {
  try {
    const data = (await response.json()) as ApiErrorResponse;
    if (data.message) {
      return data.message;
    }
  } catch {
    // 何も取得できない場合は既定文言で返す。
  }
  return fallbackMessage;
};

/**
 * API エラーメッセージを統一して返す。
 */
export const getApiErrorMessage = async (
  response: Response,
  fallbackMessage = "通信に失敗しました",
): Promise<string> => {
  const defaultMessage = statusMessageMap[response.status] ?? fallbackMessage;
  return getApiErrorMessageFromResponse(response, defaultMessage);
};
