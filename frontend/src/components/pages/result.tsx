"use client";

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

  /** おみくじ結果をXでシェアする。 */
  const handleShare = () => {
    const shareUrl = window.location.origin;
    const shareText = `#きらくじ\n\n${resultText}\n\n${shareUrl}`;
    const tweetUrl = `https://x.com/intent/tweet?text=${encodeURIComponent(shareText)}`;
    window.open(tweetUrl, "_blank", "noopener,noreferrer");
  };

  return (
    <div className="flex min-h-screen items-start justify-center bg-zinc-50 px-4 pt-[60vh] pb-10 font-sans text-zinc-900 md:px-0">
      <main className="relative flex w-full max-w-lg flex-col gap-8 rounded-none bg-zinc-900 px-6 py-10 text-center shadow-lg md:max-w-xl md:px-8 md:py-12">
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
