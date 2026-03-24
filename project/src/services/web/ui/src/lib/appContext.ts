export type AppContext = {
  csrf?: string;
  env?: string;
  Title?: string;
  error?: string;
  success?: string;
  link_uuid?: string;
  AppData?: {
    SystemLastUpdate: string;
  } | null;
  User?: {
    ID: number;
    Name: string;
    Email: string;
    Role: number;
  } | null;
  Can?: {
    Customer: boolean;
    Manager: boolean;
    Admin: boolean;
    Dev: boolean;
  };
  IsAuth?: boolean;
};

export function getAppContext(): AppContext {
  const w = window as unknown as {
    __APP_CONTEXT__?: AppContext;
  };
  return w.__APP_CONTEXT__ ?? {};
}

export function getCSRFToken(): string {
  const ctx = getAppContext();
  if (ctx.csrf) return ctx.csrf;
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta?.getAttribute("content") ?? "";
}
