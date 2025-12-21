"use client";

import Image from "next/image";
import { useRouter } from "next/navigation";
import ResultCard from "@/components/modal/result-card";

type FortuneResultProps = {
  defaultText?: string;
};

/** おみくじ結果ページの表示を組み立てる。 */
export default function FortuneResultPage({
  defaultText = "今日のきらくじ: ここに結果テキストが入ります。",
}: FortuneResultProps) {
  const router = useRouter();
  const resultText = defaultText;

  /** おみくじ結果を共有し、未対応ならコピーへ切り替える。 */
  const handleShare = async () => {
    const shareText = resultText;
    if (navigator.share) {
      try {
        await navigator.share({ text: shareText });
        return;
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") {
          return;
        }
      }
    }

    if (navigator.clipboard?.writeText) {
      try {
        await navigator.clipboard.writeText(shareText);
        window.alert("共有用のテキストをコピーしました");
        return;
      } catch (_error) {
        window.alert("共有に失敗しました");
        return;
      }
    }

    window.prompt("共有用テキストをコピーしてください。", shareText);
  };

  return (
    <div className="relative flex min-h-screen items-start justify-center overflow-hidden bg-zinc-50 px-4 pt-[60vh] pb-10 font-sans text-zinc-900 md:px-0">
      <div className="pointer-events-none absolute inset-x-0 top-0 z-0 flex justify-center opacity-70">
        <div className="w-full md:max-w-xl">
          <Image
            src="/curtain.png"
            alt=""
            width={1440}
            height={480}
            className="h-auto w-full object-top"
            priority
          />
        </div>
      </div>
      <main className="relative z-20 flex w-full max-w-lg flex-col gap-8 rounded-none bg-zinc-900 px-6 py-10 text-center shadow-lg md:max-w-xl md:px-8 md:py-12">
        <ResultCard
          resultText={resultText}
          onRetry={() => router.push("/")}
          buttonLabel="Try again"
          buttonClassName="border-zinc-200 text-zinc-100"
          secondaryButtonLabel="闇を押し付ける"
          onSecondary={handleShare}
          secondaryButtonClassName="border-zinc-200 text-zinc-100"
          reverseButtons
        />
      </main>
    </div>
  );
}
