import { useEffect, useMemo, useState } from "react";

type SaveResponse = {
  ok: boolean;
  restarted: boolean;
  message: string;
  restartError?: string;
};

type ProviderOption = {
  id: string;
  label: string;
  envKey?: string;
  group: "mainstream" | "domestic";
  supportsAutoModels: boolean;
  defaultModel?: string;
};

const providerOptions: ProviderOption[] = [
  {
    id: "openai",
    label: "OpenAI",
    envKey: "OPENAI_API_KEY",
    group: "mainstream",
    supportsAutoModels: true,
    defaultModel: "openai/gpt-4o-mini",
  },
  {
    id: "anthropic",
    label: "Anthropic",
    envKey: "ANTHROPIC_API_KEY",
    group: "mainstream",
    supportsAutoModels: true,
    defaultModel: "anthropic/claude-3-7-sonnet",
  },
  {
    id: "gemini",
    label: "Gemini",
    envKey: "GEMINI_API_KEY",
    group: "mainstream",
    supportsAutoModels: true,
    defaultModel: "gemini/gemini-1.5-pro",
  },
  {
    id: "groq",
    label: "Groq",
    envKey: "GROQ_API_KEY",
    group: "mainstream",
    supportsAutoModels: true,
    defaultModel: "groq/llama-3.1-70b-versatile",
  },
  {
    id: "mistral",
    label: "Mistral",
    envKey: "MISTRAL_API_KEY",
    group: "mainstream",
    supportsAutoModels: true,
    defaultModel: "mistral/large-latest",
  },
  {
    id: "cohere",
    label: "Cohere",
    envKey: "COHERE_API_KEY",
    group: "mainstream",
    supportsAutoModels: false,
    defaultModel: "cohere/command-r-plus",
  },
  {
    id: "minimax",
    label: "MiniMax",
    envKey: "MINIMAX_API_KEY",
    group: "domestic",
    supportsAutoModels: false,
    defaultModel: "minimax/MiniMax-M2.1",
  },
  {
    id: "deepseek",
    label: "DeepSeek",
    envKey: "DEEPSEEK_API_KEY",
    group: "domestic",
    supportsAutoModels: true,
    defaultModel: "deepseek/deepseek-chat",
  },
  {
    id: "moonshot",
    label: "Moonshot / Kimi",
    envKey: "MOONSHOT_API_KEY",
    group: "domestic",
    supportsAutoModels: true,
    defaultModel: "moonshot/kimi-k2.5",
  },
  {
    id: "zai",
    label: "ZAI / GLM",
    envKey: "ZAI_API_KEY",
    group: "domestic",
    supportsAutoModels: false,
    defaultModel: "zai/glm-4.7",
  },
  {
    id: "qwen",
    label: "Qwen",
    envKey: "QWEN_API_KEY",
    group: "domestic",
    supportsAutoModels: true,
    defaultModel: "qwen/qwen2.5-coder-32b-instruct",
  },
  {
    id: "custom",
    label: "自定义提供商",
    group: "domestic",
    supportsAutoModels: false,
  },
];

export default function App() {
  const [providerId, setProviderId] = useState("openai");
  const [providerEnvKey, setProviderEnvKey] = useState("OPENAI_API_KEY");
  const [customEnvKey, setCustomEnvKey] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [model, setModel] = useState("openai/gpt-4o-mini");
  const [models, setModels] = useState<string[]>([]);
  const [modelsLoading, setModelsLoading] = useState(false);
  const [modelsMessage, setModelsMessage] = useState<string | null>(null);
  const [gatewayToken, setGatewayToken] = useState("");
  const [status, setStatus] = useState<SaveResponse | null>(null);
  const [saving, setSaving] = useState(false);

  const canSave = useMemo(() => model.trim().length > 0, [model]);

  useEffect(() => {
    if (!gatewayToken) {
      setGatewayToken(createToken());
    }
  }, [gatewayToken]);

  useEffect(() => {
    const option = providerOptions.find((item) => item.id === providerId);
    if (providerId === "custom") {
      setProviderEnvKey(customEnvKey);
    } else {
      setProviderEnvKey(option?.envKey ?? "");
    }
    if (option?.defaultModel) {
      setModel(option.defaultModel);
    }
    setModels([]);
    setModelsMessage(null);
  }, [providerId, customEnvKey]);

  const createToken = () => {
    const bytes = crypto.getRandomValues(new Uint8Array(24));
    return Array.from(bytes)
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");
  };

  const generateToken = () => {
    setGatewayToken(createToken());
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!canSave || saving) return;
    setSaving(true);
    setStatus(null);

    try {
      const token = gatewayToken.trim() || createToken();
      if (!gatewayToken.trim()) {
        setGatewayToken(token);
      }
      const payload = {
        model: model.trim(),
        gatewayToken: token,
        providers:
          providerEnvKey.trim() && apiKey.trim()
            ? [{ key: providerEnvKey.trim(), value: apiKey.trim() }]
            : [],
      };
      const resp = await fetch("/api/config", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const data = (await resp.json()) as SaveResponse;
      setStatus(data);
    } catch (err) {
      setStatus({
        ok: false,
        restarted: false,
        message: "保存失败，请检查服务日志",
        restartError: String(err),
      });
    } finally {
      setSaving(false);
    }
  };

  const handleFetchModels = async () => {
    if (!apiKey.trim()) {
      setModelsMessage("请先填写 API Key");
      return;
    }
    setModelsLoading(true);
    setModelsMessage(null);
    try {
      const resp = await fetch("/api/models", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ provider: providerId, apiKey: apiKey.trim() }),
      });
      const data = await resp.json();
      if (!resp.ok) {
        setModelsMessage(data.message || "获取模型失败");
        setModels([]);
        return;
      }
      const list = Array.isArray(data.models) ? data.models : [];
      setModels(list);
      if (list.length === 0) {
        setModelsMessage(data.message || "未返回模型列表，可手动填写");
      } else {
        setModelsMessage(null);
      }
    } catch (err) {
      setModelsMessage("获取模型失败，请检查网络或代理");
    } finally {
      setModelsLoading(false);
    }
  };

  const handleCopyToken = async () => {
    try {
      await navigator.clipboard.writeText(gatewayToken);
      setModelsMessage("Token 已复制");
    } catch {
      setModelsMessage("复制失败，请手动复制");
    }
  };

  return (
    <div className="page">
      <div className="card">
        <header className="header">
          <h1>OpenClaw 快速配置</h1>
          <p>生成 moltbot.json 与 .env，无需执行初始化命令。</p>
        </header>

        <form className="form" onSubmit={handleSubmit}>
          <label className="field">
            <span>模型提供商</span>
            <select value={providerId} onChange={(e) => setProviderId(e.target.value)}>
              <optgroup label="主流提供商">
                {providerOptions
                  .filter((item) => item.group === "mainstream")
                  .map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.label}
                    </option>
                  ))}
              </optgroup>
              <optgroup label="国内提供商">
                {providerOptions
                  .filter((item) => item.group === "domestic")
                  .map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.label}
                    </option>
                  ))}
              </optgroup>
            </select>
          </label>

          <label className="field">
            <span>API Key</span>
            <input
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder={providerEnvKey ? `${providerEnvKey}...` : "API Key"}
            />
          </label>

          {providerId === "custom" && (
            <label className="field">
              <span>环境变量名</span>
              <input
                value={customEnvKey}
                onChange={(e) => setCustomEnvKey(e.target.value)}
                placeholder="例如 CUSTOM_API_KEY"
              />
            </label>
          )}

          <label className="field">
            <span>默认模型</span>
            <div className="inline stretch">
              <input
                value={model}
                onChange={(e) => setModel(e.target.value)}
                placeholder="如 openai/gpt-4o-mini"
                list="model-options"
                required
              />
              <button
                type="button"
                className="ghost"
                onClick={handleFetchModels}
                disabled={modelsLoading}
              >
                {modelsLoading ? "获取中" : "获取模型"}
              </button>
            </div>
            {modelsMessage && <div className="hint">{modelsMessage}</div>}
            <datalist id="model-options">
              {models.map((item) => (
                <option key={item} value={item} />
              ))}
            </datalist>
          </label>

          <label className="field">
            <span>网关 Token</span>
            <div className="inline stretch">
              <input
                className="token-input"
                value={gatewayToken}
                onChange={(e) => setGatewayToken(e.target.value)}
                placeholder="自动生成，可手动修改"
              />
              <button type="button" className="ghost" onClick={generateToken}>
                重新生成
              </button>
              <button type="button" className="ghost" onClick={handleCopyToken}>
                复制
              </button>
            </div>
          </label>

          <button type="submit" className="primary" disabled={!canSave || saving}>
            {saving ? "保存中..." : "保存并重启"}
          </button>
        </form>

        {status && (
          <div className={`status ${status.ok ? "ok" : "error"}`}>
            <strong>{status.message}</strong>
            {status.restartError && (
              <div className="status-detail">重启失败：{status.restartError}</div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
