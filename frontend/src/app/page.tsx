"use client";

import { useEffect, useState } from "react";

type Step = "input" | "loading" | "result";

export default function Home() {
  const [currentStep, setCurrentStep] = useState<Step>("input");

  useEffect(() => {
    if (currentStep !== "loading") {
      return;
    }

    const timerId = window.setTimeout(() => {
      setCurrentStep("result");
    }, 1500);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [currentStep]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 font-sans text-zinc-900">
      <main className="flex w-full max-w-xl flex-col gap-8 rounded-3xl bg-white px-8 py-12 shadow-lg">
        <header className="text-center">
          <h1 className="font-semibold text-xl">きらくじ（仮UI）</h1>
        </header>

        {currentStep === "input" && (
          <section className="flex flex-col gap-4">
            <textarea
              className="min-h-[140px] w-full resize-none rounded-2xl border border-zinc-200 px-4 py-3 text-sm outline-none"
              placeholder="ここに闇を投げる（最大140字）"
            />
            <button
              className="rounded-full bg-zinc-900 px-6 py-3 font-semibold text-sm text-white"
              type="button"
              onClick={() => setCurrentStep("loading")}
            >
              懺悔する
            </button>
          </section>
        )}

        {currentStep === "loading" && (
          <section className="flex flex-col items-center gap-3 text-center">
            <p className="font-medium text-base">
              少し待っていて、あなたのためのお告げを探すから。
            </p>
            <p className="text-sm text-zinc-500">きらくじを引いています...</p>
          </section>
        )}

        {currentStep === "result" && (
          <section className="flex flex-col gap-4 text-center">
            <p className="font-medium text-base">
              今日のきらくじ: ここに結果テキストが入ります。
            </p>
            <button
              className="rounded-full border border-zinc-300 px-6 py-3 font-semibold text-sm text-zinc-700"
              type="button"
              onClick={() => setCurrentStep("input")}
            >
              もう一度懺悔する
            </button>
          </section>
        )}
      </main>
    </div>
  );
}
