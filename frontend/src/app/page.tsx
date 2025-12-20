"use client";

import { useState } from "react";

import { fetchRandomDraw } from "../lib/draws";
import { createPost } from "../lib/posts";

type Step = "input" | "loading" | "result" | "error";

export default function Home() {
  const [currentStep, setCurrentStep] = useState<Step>("input");
  const [content, setContent] = useState("");
  const [resultText, setResultText] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const contentLength = content.length;
  const trimmedLength = content.trim().length;
  const isSubmitDisabled = trimmedLength === 0 || contentLength > 140;

  const handleSubmit = async () => {
    if (isSubmitDisabled) {
      return;
    }

    setCurrentStep("loading");
    setErrorMessage("");

    try {
      await createPost(content.trim());
      const draw = await fetchRandomDraw();
      setResultText(draw.result);
      setCurrentStep("result");
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "通信に失敗しました";
      setErrorMessage(message);
      setCurrentStep("error");
    }
  };

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
              maxLength={140}
              placeholder="ここに闇を投げる（最大140字）"
              value={content}
              onChange={(event) => setContent(event.target.value)}
            />
            <div className="text-right text-xs text-zinc-500">
              {contentLength}/140
            </div>
            <button
              className="rounded-full bg-zinc-900 px-6 py-3 font-semibold text-sm text-white disabled:cursor-not-allowed disabled:bg-zinc-400"
              type="button"
              onClick={handleSubmit}
              disabled={isSubmitDisabled}
            >
              懺悔する
            </button>
          </section>
        )}

        {currentStep === "loading" && (
          <section className="flex flex-col items-center gap-4 text-center">
            <div className="flex h-40 w-40 items-center justify-center rounded-full border border-zinc-300 bg-zinc-100 text-xs text-zinc-500">
              黒グラキャラ（仮）
            </div>
            <p className="font-medium text-base">
              少し待っていて、あなたのためのお告げを探すから。
            </p>
            <div
              className="h-2 w-full max-w-sm overflow-hidden rounded-full bg-zinc-200"
              role="progressbar"
              aria-busy="true"
              aria-label="きらくじを引いています"
            >
              <div className="h-full w-1/3 animate-pulse rounded-full bg-zinc-700" />
            </div>
            <p className="text-sm text-zinc-500">きらくじを引いています...</p>
          </section>
        )}

        {currentStep === "error" && (
          <section className="flex flex-col gap-4 text-center">
            <p className="font-medium text-base">{errorMessage}</p>
            <div className="flex flex-col gap-3">
              <button
                className="rounded-full bg-zinc-900 px-6 py-3 font-semibold text-sm text-white"
                type="button"
                onClick={handleSubmit}
              >
                もう一度試す
              </button>
              <button
                className="rounded-full border border-zinc-300 px-6 py-3 font-semibold text-sm text-zinc-700"
                type="button"
                onClick={() => setCurrentStep("input")}
              >
                入力に戻る
              </button>
            </div>
          </section>
        )}

        {currentStep === "result" && (
          <section className="flex flex-col gap-4 text-center">
            <p className="font-medium text-base">{resultText}</p>
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
