package sdk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

const (
	compactionPromptFileEnv = "PI_GOLANG_COMPACTION_PROMPT_FILE"
	compactionPromptHashEnv = "PI_GOLANG_COMPACTION_PROMPT_SHA256"
)

type managedCompactionHook struct {
	bundleDir     string
	extensionPath string
	promptPath    string
	promptHash    string
}

func createManagedCompactionHook(prompt string) (*managedCompactionHook, error) {
	bundleDir, err := os.MkdirTemp("", "pi-golang-compaction-hook-")
	if err != nil {
		return nil, fmt.Errorf("create compaction hook temp dir: %w", err)
	}

	promptPath := filepath.Join(bundleDir, "compaction.prompt")
	if err := os.WriteFile(promptPath, []byte(prompt), 0o600); err != nil {
		_ = os.RemoveAll(bundleDir)
		return nil, fmt.Errorf("write compaction prompt file: %w", err)
	}

	extensionPath := filepath.Join(bundleDir, "compaction-hook.ts")
	extensionSource := renderCompactionHookExtension()
	if err := os.WriteFile(extensionPath, []byte(extensionSource), 0o600); err != nil {
		_ = os.RemoveAll(bundleDir)
		return nil, fmt.Errorf("write compaction hook extension: %w", err)
	}

	sum := sha256.Sum256([]byte(prompt))
	promptHash := hex.EncodeToString(sum[:])

	return &managedCompactionHook{
		bundleDir:     bundleDir,
		extensionPath: extensionPath,
		promptPath:    promptPath,
		promptHash:    promptHash,
	}, nil
}

func (hook *managedCompactionHook) arguments() []string {
	if hook == nil {
		return nil
	}
	return []string{"--extension", hook.extensionPath}
}

func (hook *managedCompactionHook) injectEnvironment(values map[string]string) {
	if hook == nil {
		return
	}
	values[compactionPromptFileEnv] = hook.promptPath
	values[compactionPromptHashEnv] = hook.promptHash
}

func (hook *managedCompactionHook) cleanup() error {
	if hook == nil || hook.bundleDir == "" {
		return nil
	}
	return os.RemoveAll(hook.bundleDir)
}

func renderCompactionHookExtension() string {
	return fmt.Sprintf(`import { readFileSync } from "node:fs";
import { complete } from "@mariozechner/pi-ai";
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { convertToLlm, serializeConversation } from "@mariozechner/pi-coding-agent";

const PROMPT_FILE_ENV = %q;
const PROMPT_HASH_ENV = %q;
const SOURCE = "pi-golang-compaction-prompt";

function readCompactionPrompt(): string | undefined {
	const promptPath = process.env[PROMPT_FILE_ENV];
	if (!promptPath) return undefined;
	try {
		const prompt = readFileSync(promptPath, "utf-8");
		return prompt.trim().length > 0 ? prompt : undefined;
	} catch {
		return undefined;
	}
}

export default function (pi: ExtensionAPI) {
	const compactionPrompt = readCompactionPrompt();
	const promptHash = process.env[PROMPT_HASH_ENV] || "";
	if (!compactionPrompt) return;

	pi.on("session_before_compact", async (event, ctx) => {
		const model = ctx.model;
		if (!model) return;

		const apiKey = await ctx.modelRegistry.getApiKey(model);
		if (!apiKey) return;

		const prep = event.preparation;
		const messages = [...prep.messagesToSummarize, ...prep.turnPrefixMessages];
		const conversationText = serializeConversation(convertToLlm(messages));

		const sections = [compactionPrompt.trim()];
		if (event.customInstructions?.trim()) {
			sections.push("Additional compaction instructions:\n" + event.customInstructions.trim());
		}
		if (prep.previousSummary?.trim()) {
			sections.push("<previous-summary>\n" + prep.previousSummary.trim() + "\n</previous-summary>");
		}
		sections.push("<conversation>\n" + conversationText + "\n</conversation>");

		const promptText = sections.join("\n\n");
		const summarizationMessages = [
			{
				role: "user" as const,
				content: [{ type: "text" as const, text: promptText }],
				timestamp: Date.now(),
			},
		];

		const reserveTokens = prep.settings?.reserveTokens ?? 16384;
		const maxTokens = Math.max(512, Math.floor(reserveTokens * 0.75));

		try {
			const response = await complete(model, { messages: summarizationMessages }, { apiKey, maxTokens, signal: event.signal });
			if (response.stopReason === "error") return;

			const summary = response.content
				.filter((content): content is { type: "text"; text: string } => content.type === "text")
				.map((content) => content.text)
				.join("\n")
				.trim();
			if (!summary) return;

			return {
				compaction: {
					summary,
					firstKeptEntryId: prep.firstKeptEntryId,
					tokensBefore: prep.tokensBefore,
					details: {
						source: SOURCE,
						promptHash,
					},
				},
			};
		} catch {
			return;
		}
	});
}
`, compactionPromptFileEnv, compactionPromptHashEnv)
}
