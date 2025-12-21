import FortuneResultPage from "@/components/pages/result";

type ResultPageProps = {
  searchParams?: Promise<{
    text?: string | string[];
  }>;
};

/**
 * おみくじ結果ページを描画する。
 */
export default async function ResultPage({ searchParams }: ResultPageProps) {
  const resolvedSearchParams = searchParams ? await searchParams : undefined;
  const textParam = resolvedSearchParams?.text;
  const defaultText =
    typeof textParam === "string" && textParam.trim() !== ""
      ? textParam
      : undefined;

  return <FortuneResultPage defaultText={defaultText} />;
}
