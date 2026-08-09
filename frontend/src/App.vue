<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import {
  CopyPassword,
  CreateAccount,
  DeleteAccount,
  GeneratePassword,
  GetAccount,
  GetLockState,
  GetSettings,
  InitializeVault,
  ListAccounts,
  Lock,
  RecordActivity,
  SearchAccounts,
  Unlock,
  UpdateAccount,
  UpdateSettings,
} from '../wailsjs/go/app/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

type Language = 'zh' | 'en';
type ViewMode = 'accounts' | 'generator' | 'settings';
type EditorMode = 'none' | 'new' | 'edit';

type LockState = {
  initialized: boolean;
  locked: boolean;
  auto_lock_minutes?: number;
};

type AccountSummary = {
  id: string;
  title: string;
  username: string;
  url: string;
  created_at: number;
  updated_at: number;
};

type AccountDetail = AccountSummary & {
  password: string;
  note: string;
};

type AccountInput = {
  title: string;
  username: string;
  password: string;
  url: string;
  note: string;
};

const messages = {
  en: {
    brandSubtitle: 'KEy VAult',
    languageButton: '中文',
    navAccounts: 'Accounts',
    navGenerator: 'Password generator',
    navSettings: 'Settings',
    firstRun: 'First run',
    createVault: 'Create the local vault',
    createVaultBody: 'Choose a Master Password to create your encrypted local vault.',
    initializeVault: 'Initialize vault',
    lockedEyebrow: 'Locked',
    lockedTitle: 'Vault locked',
    lockedBody: 'Your vault is sealed. Enter your Master Password to continue.',
    lockedWarning: 'KEVA is locked. Account details and passwords are unavailable until you unlock the vault.',
    unlock: 'Unlock',
    unlocked: 'Unlocked',
    workspace: 'Vault workspace',
    accountCount: (count: number) => `${count} account${count === 1 ? '' : 's'}`,
    lockVault: 'Lock vault',
    masterPassword: 'Master Password',
    working: 'Working...',
    searchAccounts: 'Search accounts',
    new: 'New',
    noAccounts: 'No accounts yet.',
    emptyTitle: 'No account selected',
    emptyBody: 'Select an account to view details, or create a new one.',
    newAccount: 'New account',
    editAccount: 'Edit account',
    untitled: 'Untitled entry',
    updated: 'Updated',
    title: 'Title',
    username: 'Username',
    password: 'Password',
    show: 'Show',
    hide: 'Hide',
    generate: 'Generate',
    copy: 'Copy',
    copied: 'Copied',
    url: 'URL',
    note: 'Note',
    required: 'Required',
    optional: 'Optional',
    titlePlaceholder: 'Website name, app name, or service name',
    usernamePlaceholder: 'Username, phone number, or email',
    passwordPlaceholder: 'Password, passkey, API key, or secret',
    urlPlaceholder: 'Website URL or login page',
    notePlaceholder: 'Notes, description, purpose, or recovery hints',
    saveChanges: 'Save changes',
    createAccount: 'Create account',
    cancel: 'Cancel',
    delete: 'Delete',
    confirmDelete: (title: string) => `Delete "${title || 'this account'}"?`,
    generatorTitle: 'Password generator',
    length: 'Length',
    uppercase: 'Uppercase',
    lowercase: 'Lowercase',
    numbers: 'Numbers',
    symbols: 'Symbols',
    generatedPlaceholder: 'Generated password appears here',
    generatePassword: 'Generate password',
    useInAccount: 'Use in account',
    settings: 'Settings',
    autoLockMinutes: 'Auto-lock minutes',
    saveSettings: 'Save settings',
    checking: 'Checking vault state...',
    requiredMasterPassword: 'Master Password is required.',
    vaultInitialized: 'Vault initialized.',
    vaultUnlocked: 'Vault unlocked.',
    vaultLocked: 'Vault locked.',
    autoLocked: 'Vault locked after inactivity.',
    accountUpdated: 'Account updated.',
    accountCreated: 'Account created.',
    accountDeleted: 'Account deleted.',
    passwordGenerated: 'Password generated.',
    passwordAdded: 'Password added to the form.',
    passwordCopied: 'Password copied.',
    settingsSaved: 'Settings saved.',
    noUsername: 'No username',
    errors: {
      'authentication failed': 'authentication failed',
    },
  },
  zh: {
    brandSubtitle: '账号密码保险库',
    languageButton: 'EN',
    navAccounts: '账号',
    navGenerator: '密码生成器',
    navSettings: '设置',
    firstRun: '首次运行',
    createVault: '创建本地保险库',
    createVaultBody: '设置密钥，创建你的本地加密保险库。',
    initializeVault: '初始化保险库',
    lockedEyebrow: '已锁定',
    lockedTitle: '保险库已锁定',
    lockedBody: '你的保险库已锁定。输入密钥后继续使用。',
    lockedWarning: 'KEVA 当前处于锁定状态。解锁前无法查看账号详情和密码。',
    unlock: '解锁',
    unlocked: '已解锁',
    workspace: '保险库工作区',
    accountCount: (count: number) => `${count} 个账号`,
    lockVault: '锁定保险库',
    masterPassword: '密钥',
    working: '处理中...',
    searchAccounts: '搜索账号',
    new: '新增',
    noAccounts: '还没有账号。',
    emptyTitle: '未选择账号',
    emptyBody: '选择一个账号查看详情，或点击新增创建账号。',
    newAccount: '新增账号',
    editAccount: '编辑账号',
    untitled: '未命名条目',
    updated: '更新于',
    title: '标题',
    username: '用户名',
    password: '密码',
    show: '显示',
    hide: '隐藏',
    generate: '生成',
    copy: '复制',
    copied: '已复制',
    url: 'URL',
    note: '备注',
    required: '必填',
    optional: '选填',
    titlePlaceholder: '可以输入网站名称/应用名称等等',
    usernamePlaceholder: '可以输入用户名/手机号/邮箱等等',
    passwordPlaceholder: '可以输入密码/密钥等等',
    urlPlaceholder: '可以输入网站地址/登录页面等等',
    notePlaceholder: '可以输入备注信息/描述/用途等等',
    saveChanges: '保存修改',
    createAccount: '创建账号',
    cancel: '取消',
    delete: '删除',
    confirmDelete: (title: string) => `删除“${title || '这个账号'}”？`,
    generatorTitle: '密码生成器',
    length: '长度',
    uppercase: '大写字母',
    lowercase: '小写字母',
    numbers: '数字',
    symbols: '符号',
    generatedPlaceholder: '生成的密码会显示在这里',
    generatePassword: '生成密码',
    useInAccount: '填入账号',
    settings: '设置',
    autoLockMinutes: '自动锁定分钟数',
    saveSettings: '保存设置',
    checking: '正在检查保险库状态...',
    requiredMasterPassword: '请输入密钥。',
    vaultInitialized: '保险库已初始化。',
    vaultUnlocked: '保险库已解锁。',
    vaultLocked: '保险库已锁定。',
    autoLocked: '长时间未操作，保险库已自动锁定。',
    accountUpdated: '账号已更新。',
    accountCreated: '账号已创建。',
    accountDeleted: '账号已删除。',
    passwordGenerated: '密码已生成。',
    passwordAdded: '密码已填入表单。',
    passwordCopied: '密码已复制。',
    settingsSaved: '设置已保存。',
    noUsername: '无用户名',
    errors: {
      'authentication failed': '认证失败，请检查密钥后重试。',
    },
  },
};

const emptyAccount = (): AccountInput => ({
  title: '',
  username: '',
  password: '',
  url: '',
  note: '',
});

const initialLanguage = (): Language => {
  const stored = window.localStorage.getItem('keva_language');
  if (stored === 'zh' || stored === 'en') return stored;
  return navigator.language.toLowerCase().startsWith('zh') ? 'zh' : 'en';
};

const language = ref<Language>(initialLanguage());
const t = computed(() => messages[language.value]);
const loading = ref(true);
const busy = ref(false);
const error = ref('');
const notice = ref('');
const masterPassword = ref('');
const searchInput = ref<HTMLInputElement | null>(null);
const search = ref('');
const accounts = ref<AccountSummary[]>([]);
const selected = ref<AccountDetail | null>(null);
const editingId = ref('');
const activeView = ref<ViewMode>('accounts');
const editorMode = ref<EditorMode>('none');
const revealPassword = ref(false);
const copiedAccountId = ref('');
const generatedPassword = ref('');
const accountForm = reactive<AccountInput>(emptyAccount());
const settings = reactive({
  auto_lock_minutes: 5,
});
const passwordOptions = reactive({
  length: 20,
  include_uppercase: true,
  include_lowercase: true,
  include_numbers: true,
  include_symbols: true,
});
const state = ref<LockState>({
  initialized: false,
  locked: true,
});
let activityTimer = 0;
let removeAutoLockedListener: (() => void) | undefined;
let removeShortcutListener: (() => void) | undefined;

const currentMode = computed(() => {
  if (!state.value.initialized) {
    return {
      eyebrow: t.value.firstRun,
      title: t.value.createVault,
      body: t.value.createVaultBody,
      action: t.value.initializeVault,
      status: t.value.firstRun,
    };
  }

  if (state.value.locked) {
    return {
      eyebrow: t.value.lockedEyebrow,
      title: t.value.lockedTitle,
      body: t.value.lockedBody,
      action: t.value.unlock,
      status: t.value.lockedEyebrow,
    };
  }

  return {
    eyebrow: t.value.unlocked,
    title: t.value.workspace,
    body: t.value.accountCount(accounts.value.length),
    action: '',
    status: t.value.unlocked,
  };
});

const selectedSummary = computed(() => accounts.value.find((account) => account.id === editingId.value));
const isLockedView = computed(() => state.value.initialized && state.value.locked);
const isEditorVisible = computed(() => editorMode.value !== 'none');

onMounted(async () => {
  installRuntimeListeners();
  installActivityListeners();
  await refreshState();
});

onUnmounted(() => {
  removeAutoLockedListener?.();
  removeShortcutListener?.();
  window.removeEventListener('mousemove', queueActivity);
  window.removeEventListener('keydown', queueActivity);
  window.removeEventListener('pointerdown', queueActivity);
  window.removeEventListener('scroll', queueActivity, true);
});

async function refreshState() {
  try {
    state.value = await GetLockState();
    if (state.value.auto_lock_minutes) {
      settings.auto_lock_minutes = state.value.auto_lock_minutes;
    }
    if (state.value.initialized && !state.value.locked) {
      await Promise.all([loadAccounts(), loadSettings()]);
    } else {
      accounts.value = [];
      selected.value = null;
      resetForm();
    }
  } catch (err) {
    showError(err);
  } finally {
    loading.value = false;
  }
}

async function loadAccounts() {
  queueActivity();
  accounts.value = search.value.trim() ? await SearchAccounts(search.value) : await ListAccounts();
}

async function loadSettings() {
  const next = await GetSettings();
  settings.auto_lock_minutes = next.auto_lock_minutes;
}

function toggleLanguage() {
  language.value = language.value === 'zh' ? 'en' : 'zh';
  window.localStorage.setItem('keva_language', language.value);
}

async function submitMasterPassword() {
  clearMessages();

  if (!masterPassword.value) {
    error.value = t.value.requiredMasterPassword;
    return;
  }

  busy.value = true;
  try {
    if (!state.value.initialized) {
      await InitializeVault({ master_password: masterPassword.value });
      notice.value = t.value.vaultInitialized;
    } else {
      await Unlock({ master_password: masterPassword.value });
      notice.value = t.value.vaultUnlocked;
    }
    activeView.value = 'accounts';
    resetForm();
    await refreshState();
  } catch (err) {
    showError(err);
  } finally {
    masterPassword.value = '';
    busy.value = false;
  }
}

async function lockVault() {
  clearMessages();
  busy.value = true;
  try {
    await Lock();
    notice.value = t.value.vaultLocked;
    await refreshState();
  } catch (err) {
    showError(err);
  } finally {
    busy.value = false;
  }
}

function openNewAccount() {
  queueActivity();
  clearMessages();
  editingId.value = '';
  selected.value = null;
  revealPassword.value = false;
  editorMode.value = 'new';
  Object.assign(accountForm, emptyAccount());
}

async function selectAccount(id: string) {
  queueActivity();
  clearMessages();
  try {
    selected.value = await GetAccount(id);
    Object.assign(accountForm, {
      title: selected.value.title,
      username: selected.value.username,
      password: selected.value.password,
      url: selected.value.url,
      note: selected.value.note,
    });
    editingId.value = id;
    editorMode.value = 'edit';
    revealPassword.value = false;
  } catch (err) {
    showError(err);
  }
}

async function saveAccount() {
  queueActivity();
  clearMessages();
  busy.value = true;
  try {
    const saved = editingId.value
      ? await UpdateAccount(editingId.value, accountForm)
      : await CreateAccount(accountForm);
    notice.value = editingId.value ? t.value.accountUpdated : t.value.accountCreated;
    await loadAccounts();
    await selectAccount(saved.id);
  } catch (err) {
    showError(err);
  } finally {
    busy.value = false;
  }
}

async function removeAccount() {
  queueActivity();
  if (!editingId.value) return;
  if (!window.confirm(t.value.confirmDelete(accountForm.title))) return;

  clearMessages();
  busy.value = true;
  try {
    await DeleteAccount(editingId.value);
    notice.value = t.value.accountDeleted;
    resetForm();
    await loadAccounts();
  } catch (err) {
    showError(err);
  } finally {
    busy.value = false;
  }
}

async function generateIntoForm() {
  queueActivity();
  clearMessages();
  try {
    accountForm.password = await GeneratePassword(passwordOptions);
    revealPassword.value = true;
    notice.value = t.value.passwordGenerated;
  } catch (err) {
    showError(err);
  }
}

async function generateStandalonePassword() {
  queueActivity();
  clearMessages();
  try {
    generatedPassword.value = await GeneratePassword(passwordOptions);
    notice.value = t.value.passwordGenerated;
  } catch (err) {
    showError(err);
  }
}

function useGeneratedPassword() {
  queueActivity();
  if (!generatedPassword.value) return;
  accountForm.password = generatedPassword.value;
  activeView.value = 'accounts';
  if (editorMode.value === 'none') {
    openNewAccount();
  }
  revealPassword.value = true;
  notice.value = t.value.passwordAdded;
}

async function copySelectedPassword() {
  queueActivity();
  if (!editingId.value) return;

  clearMessages();
  try {
    await CopyPassword(editingId.value);
    copiedAccountId.value = editingId.value;
    notice.value = t.value.passwordCopied;
    window.setTimeout(() => {
      if (copiedAccountId.value === editingId.value) {
        copiedAccountId.value = '';
      }
    }, 2500);
  } catch (err) {
    showError(err);
  }
}

async function saveSettings() {
  queueActivity();
  clearMessages();
  busy.value = true;
  try {
    await UpdateSettings({ auto_lock_minutes: settings.auto_lock_minutes });
    notice.value = t.value.settingsSaved;
  } catch (err) {
    showError(err);
  } finally {
    busy.value = false;
  }
}

function resetForm() {
  editingId.value = '';
  selected.value = null;
  editorMode.value = 'none';
  revealPassword.value = false;
  Object.assign(accountForm, emptyAccount());
}

function clearMessages() {
  error.value = '';
  notice.value = '';
}

function showError(err: unknown) {
  const message = err instanceof Error ? err.message : String(err);
  const normalized = message.trim().toLowerCase();
  const errors = t.value.errors as Record<string, string>;
  const translated = errors[normalized] || (normalized.includes('authentication failed') ? errors['authentication failed'] : '');
  error.value = translated || message;
}

function installRuntimeListeners() {
  removeAutoLockedListener = EventsOn('vault:auto-locked', async () => {
    notice.value = t.value.autoLocked;
    await refreshState();
  });
  removeShortcutListener = EventsOn('shortcut:quick-search', async () => {
    activeView.value = 'accounts';
    if (!state.value.locked) {
      await loadAccounts();
      window.setTimeout(() => searchInput.value?.focus(), 50);
    }
  });
}

function installActivityListeners() {
  window.addEventListener('mousemove', queueActivity);
  window.addEventListener('keydown', queueActivity);
  window.addEventListener('pointerdown', queueActivity);
  window.addEventListener('scroll', queueActivity, true);
}

function queueActivity() {
  if (state.value.locked || !state.value.initialized) return;
  if (activityTimer) return;

  activityTimer = window.setTimeout(async () => {
    activityTimer = 0;
    try {
      await RecordActivity();
    } catch {
      // Activity pings must never interrupt the user's current workflow.
    }
  }, 1000);
}
</script>

<template>
  <main class="min-h-screen bg-[#f3f8fb] text-[#24313a]">
    <section class="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-1 gap-6 px-5 py-6 lg:grid-cols-[260px_1fr] lg:px-8">
      <aside class="flex flex-col justify-between border-b border-[#cfdee7] pb-5 lg:border-b-0 lg:border-r lg:pb-0 lg:pr-6">
        <div>
          <div class="mb-7 flex items-center gap-3">
            <div class="grid h-11 w-11 place-items-center border border-[#159c8b] bg-white text-sm font-semibold text-[#0d7f73] shadow-sm shadow-[#8abeb5]/30">
              KV
            </div>
            <div>
              <p class="text-sm uppercase tracking-[0.22em] text-[#6c808c]">{{ t.brandSubtitle }}</p>
              <h1 class="text-2xl font-semibold text-[#17232c]">KEVA</h1>
            </div>
          </div>

          <nav class="space-y-2 text-sm text-[#536872]">
            <button
              class="block w-full px-4 py-3 text-left"
              :class="activeView === 'accounts' ? 'border-l-2 border-[#159c8b] bg-white text-[#17232c] shadow-sm' : ''"
              type="button"
              @click="activeView = 'accounts'"
            >
              {{ t.navAccounts }}
            </button>
            <button
              class="block w-full px-4 py-3 text-left disabled:opacity-40"
              :class="activeView === 'generator' ? 'border-l-2 border-[#159c8b] bg-white text-[#17232c] shadow-sm' : ''"
              type="button"
              :disabled="state.locked"
              @click="activeView = 'generator'"
            >
              {{ t.navGenerator }}
            </button>
            <button
              class="block w-full px-4 py-3 text-left disabled:opacity-40"
              :class="activeView === 'settings' ? 'border-l-2 border-[#159c8b] bg-white text-[#17232c] shadow-sm' : ''"
              type="button"
              :disabled="state.locked"
              @click="activeView = 'settings'"
            >
              {{ t.navSettings }}
            </button>
          </nav>
        </div>

        <div class="mt-8 grid gap-3 text-xs leading-5 text-[#6c808c]">
          <div
            class="border bg-white p-3"
            :class="isLockedView ? 'border-[#e28a8a] bg-[#fff1f1] text-[#9b1c1c]' : 'border-[#cfdee7]'"
          >
            <p class="font-semibold" :class="isLockedView ? 'text-[#9b1c1c]' : 'text-[#17232c]'">{{ currentMode.status }}</p>
            <p v-if="isLockedView">{{ t.lockedWarning }}</p>
          </div>
        </div>
      </aside>

      <section class="w-full border border-[#cfdee7] bg-white p-5 shadow-2xl shadow-[#9fb8c8]/25 sm:p-6">
        <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
          <div>
            <p class="text-sm uppercase tracking-[0.22em]" :class="isLockedView ? 'text-[#b42318]' : 'text-[#0d7f73]'">
              {{ currentMode.eyebrow }}
            </p>
            <h2 class="mt-2 text-3xl font-semibold sm:text-4xl" :class="isLockedView ? 'text-[#9b1c1c]' : 'text-[#17232c]'">
              {{ currentMode.title }}
            </h2>
          </div>
          <div class="flex items-center gap-3">
            <button class="border border-[#b7cbd6] bg-white px-4 py-2 text-sm font-semibold text-[#40535d]" type="button" @click="toggleLanguage">
              {{ t.languageButton }}
            </button>
            <button
              v-if="state.initialized && !state.locked"
              class="border border-[#b42318] bg-[#fff1f1] px-4 py-2 text-sm font-semibold text-[#9b1c1c]"
              type="button"
              @click="lockVault"
            >
              {{ t.lockVault }}
            </button>
          </div>
        </div>

        <div v-if="loading" class="border border-[#cfdee7] bg-[#f7fbfd] p-5 text-[#536872]">
          {{ t.checking }}
        </div>

        <div
          v-else-if="state.locked || !state.initialized"
          class="relative overflow-hidden border p-6"
          :class="isLockedView ? 'border-[#e28a8a] bg-[#fff1f1]' : 'border-[#cfdee7] bg-[#f7fbfd]'"
        >
          <div class="absolute left-0 top-0 h-full w-2" :class="isLockedView ? 'bg-[#d92d20]' : 'bg-[#159c8b]'"></div>
          <div class="pl-4">
            <p class="max-w-2xl text-base leading-7" :class="isLockedView ? 'text-[#7a271a]' : 'text-[#40535d]'">
              {{ currentMode.body }}
            </p>
            <p v-if="isLockedView" class="mt-3 max-w-2xl text-sm font-semibold text-[#b42318]">
              {{ t.lockedWarning }}
            </p>

            <form class="mt-6 grid max-w-md gap-3" @submit.prevent="submitMasterPassword">
              <label class="text-sm font-semibold" :class="isLockedView ? 'text-[#7a271a]' : 'text-[#40535d]'" for="master-password">
                {{ t.masterPassword }}
              </label>
              <input
                id="master-password"
                v-model="masterPassword"
                class="border bg-white px-4 py-3 text-[#17232c] outline-none transition"
                :class="isLockedView ? 'border-[#e28a8a] focus:border-[#d92d20]' : 'border-[#b7cbd6] focus:border-[#159c8b]'"
                type="password"
                autocomplete="current-password"
              />
              <button
                class="w-fit border px-5 py-3 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-60"
                :class="isLockedView ? 'border-[#b42318] bg-[#d92d20] text-white hover:bg-[#b42318]' : 'border-[#159c8b] bg-[#dff5f1] text-[#0d5f57] hover:bg-[#c9eee8]'"
                type="submit"
                :disabled="busy"
              >
                {{ busy ? t.working : currentMode.action }}
              </button>
            </form>
          </div>
        </div>

        <div v-else-if="activeView === 'accounts'" class="grid gap-5 xl:grid-cols-[360px_1fr]">
          <section class="border border-[#cfdee7] bg-[#f7fbfd]">
            <div class="border-b border-[#cfdee7] p-4">
              <div class="flex items-center justify-between gap-3">
                <input
                  ref="searchInput"
                  v-model="search"
                  class="min-w-0 flex-1 border border-[#b7cbd6] bg-white px-3 py-2 text-sm outline-none focus:border-[#159c8b]"
                  :placeholder="t.searchAccounts"
                  type="search"
                  @input="loadAccounts"
                />
                <button class="border border-[#159c8b] bg-[#dff5f1] px-3 py-2 text-sm font-semibold text-[#0d5f57]" type="button" @click="openNewAccount">
                  {{ t.new }}
                </button>
              </div>
            </div>

            <div class="max-h-[560px] overflow-auto">
              <button
                v-for="account in accounts"
                :key="account.id"
                class="block w-full border-b border-[#dce8ee] px-4 py-3 text-left hover:bg-white"
                :class="account.id === editingId ? 'bg-white' : ''"
                type="button"
                @click="selectAccount(account.id)"
              >
                <p class="font-semibold text-[#17232c]">{{ account.title }}</p>
                <p class="mt-1 truncate text-sm text-[#6c808c]">{{ account.username || account.url || t.noUsername }}</p>
              </button>
              <div v-if="accounts.length === 0" class="p-5 text-sm text-[#6c808c]">
                {{ t.noAccounts }}
              </div>
            </div>
          </section>

          <section v-if="!isEditorVisible" class="grid min-h-[360px] place-items-center border border-[#cfdee7] bg-[#f7fbfd] p-5 text-center">
            <div class="max-w-sm">
              <p class="text-sm uppercase tracking-[0.18em] text-[#0d7f73]">{{ t.navAccounts }}</p>
              <h3 class="mt-2 text-2xl font-semibold text-[#17232c]">{{ t.emptyTitle }}</h3>
              <p class="mt-3 text-sm leading-6 text-[#6c808c]">{{ t.emptyBody }}</p>
              <button class="mt-5 border border-[#159c8b] bg-[#dff5f1] px-5 py-3 text-sm font-semibold text-[#0d5f57]" type="button" @click="openNewAccount">
                {{ t.new }}
              </button>
            </div>
          </section>

          <section v-else class="border border-[#cfdee7] bg-[#f7fbfd] p-5">
            <div class="mb-5 flex flex-wrap items-center justify-between gap-3">
              <div>
                <p class="text-sm uppercase tracking-[0.18em] text-[#0d7f73]">{{ editorMode === 'edit' ? t.editAccount : t.newAccount }}</p>
                <h3 class="mt-2 text-2xl font-semibold text-[#17232c]">{{ accountForm.title || t.untitled }}</h3>
                <p v-if="selectedSummary" class="mt-1 text-sm text-[#6c808c]">{{ t.updated }} {{ new Date(selectedSummary.updated_at * 1000).toLocaleString() }}</p>
              </div>
            </div>

            <form class="grid gap-4" @submit.prevent="saveAccount">
              <div class="grid gap-4 md:grid-cols-2">
                <label class="grid gap-2 text-sm text-[#40535d]">
                  <span class="flex items-center gap-2">
                    <span>{{ t.title }}</span>
                    <span class="border border-[#159c8b] bg-[#dff5f1] px-2 py-0.5 text-xs font-semibold text-[#0d5f57]">{{ t.required }}</span>
                  </span>
                  <input
                    v-model="accountForm.title"
                    class="border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]"
                    :placeholder="t.titlePlaceholder"
                  />
                </label>
                <label class="grid gap-2 text-sm text-[#40535d]">
                  <span class="flex items-center gap-2">
                    <span>{{ t.username }}</span>
                    <span class="border border-[#cfdee7] bg-white px-2 py-0.5 text-xs font-semibold text-[#6c808c]">{{ t.optional }}</span>
                  </span>
                  <input
                    v-model="accountForm.username"
                    class="border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]"
                    :placeholder="t.usernamePlaceholder"
                  />
                </label>
              </div>

              <label class="grid gap-2 text-sm text-[#40535d]">
                <span class="flex items-center gap-2">
                  <span>{{ t.password }}</span>
                  <span class="border border-[#159c8b] bg-[#dff5f1] px-2 py-0.5 text-xs font-semibold text-[#0d5f57]">{{ t.required }}</span>
                </span>
                <div class="grid gap-2 sm:grid-cols-[1fr_auto_auto_auto]">
                  <input
                    v-model="accountForm.password"
                    class="border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]"
                    :placeholder="t.passwordPlaceholder"
                    :type="revealPassword ? 'text' : 'password'"
                  />
                  <button class="border border-[#b7cbd6] bg-white px-3 py-2 text-sm text-[#40535d]" type="button" @click="revealPassword = !revealPassword">
                    {{ revealPassword ? t.hide : t.show }}
                  </button>
                  <button class="border border-[#b7cbd6] bg-white px-3 py-2 text-sm text-[#40535d]" type="button" @click="generateIntoForm">
                    {{ t.generate }}
                  </button>
                  <button class="border border-[#b7cbd6] bg-white px-3 py-2 text-sm text-[#40535d] disabled:opacity-50" type="button" :disabled="!editingId" @click="copySelectedPassword">
                    {{ copiedAccountId === editingId ? t.copied : t.copy }}
                  </button>
                </div>
              </label>

              <label class="grid gap-2 text-sm text-[#40535d]">
                <span class="flex items-center gap-2">
                  <span>{{ t.url }}</span>
                  <span class="border border-[#cfdee7] bg-white px-2 py-0.5 text-xs font-semibold text-[#6c808c]">{{ t.optional }}</span>
                </span>
                <input
                  v-model="accountForm.url"
                  class="border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]"
                  :placeholder="t.urlPlaceholder"
                />
              </label>

              <label class="grid gap-2 text-sm text-[#40535d]">
                <span class="flex items-center gap-2">
                  <span>{{ t.note }}</span>
                  <span class="border border-[#cfdee7] bg-white px-2 py-0.5 text-xs font-semibold text-[#6c808c]">{{ t.optional }}</span>
                </span>
                <textarea
                  v-model="accountForm.note"
                  class="min-h-24 border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]"
                  :placeholder="t.notePlaceholder"
                ></textarea>
              </label>

              <div class="flex flex-wrap gap-3">
                <button class="border border-[#159c8b] bg-[#dff5f1] px-5 py-3 text-sm font-semibold text-[#0d5f57] disabled:opacity-60" type="submit" :disabled="busy">
                  {{ editingId ? t.saveChanges : t.createAccount }}
                </button>
                <button class="border border-[#b7cbd6] bg-white px-5 py-3 text-sm font-semibold text-[#40535d]" type="button" @click="resetForm">
                  {{ t.cancel }}
                </button>
                <button class="border border-[#d98b92] bg-[#fff1f2] px-5 py-3 text-sm font-semibold text-[#8e333b] disabled:opacity-50" type="button" :disabled="!editingId || busy" @click="removeAccount">
                  {{ t.delete }}
                </button>
              </div>
            </form>
          </section>
        </div>

        <div v-else-if="activeView === 'generator'" class="grid gap-5 xl:grid-cols-[360px_1fr]">
          <section class="border border-[#cfdee7] bg-[#f7fbfd] p-5">
            <h3 class="text-2xl font-semibold text-[#17232c]">{{ t.generatorTitle }}</h3>
            <div class="mt-5 grid gap-5">
              <label class="grid gap-2 text-sm text-[#40535d]">
                {{ t.length }}: {{ passwordOptions.length }}
                <input v-model.number="passwordOptions.length" class="accent-[#159c8b]" max="128" min="8" type="range" />
              </label>
              <label class="flex items-center gap-3 text-sm text-[#40535d]">
                <input v-model="passwordOptions.include_uppercase" class="accent-[#159c8b]" type="checkbox" />
                {{ t.uppercase }}
              </label>
              <label class="flex items-center gap-3 text-sm text-[#40535d]">
                <input v-model="passwordOptions.include_lowercase" class="accent-[#159c8b]" type="checkbox" />
                {{ t.lowercase }}
              </label>
              <label class="flex items-center gap-3 text-sm text-[#40535d]">
                <input v-model="passwordOptions.include_numbers" class="accent-[#159c8b]" type="checkbox" />
                {{ t.numbers }}
              </label>
              <label class="flex items-center gap-3 text-sm text-[#40535d]">
                <input v-model="passwordOptions.include_symbols" class="accent-[#159c8b]" type="checkbox" />
                {{ t.symbols }}
              </label>
            </div>
          </section>

          <section class="border border-[#cfdee7] bg-[#f7fbfd] p-5">
            <div class="break-all border border-[#b7cbd6] bg-white p-4 font-mono text-sm text-[#17232c]">
              {{ generatedPassword || t.generatedPlaceholder }}
            </div>
            <div class="mt-4 flex flex-wrap gap-3">
              <button class="border border-[#159c8b] bg-[#dff5f1] px-5 py-3 text-sm font-semibold text-[#0d5f57]" type="button" @click="generateStandalonePassword">
                {{ t.generatePassword }}
              </button>
              <button class="border border-[#b7cbd6] bg-white px-5 py-3 text-sm font-semibold text-[#40535d] disabled:opacity-50" type="button" :disabled="!generatedPassword" @click="useGeneratedPassword">
                {{ t.useInAccount }}
              </button>
            </div>
          </section>
        </div>

        <div v-else class="max-w-xl border border-[#cfdee7] bg-[#f7fbfd] p-5">
          <h3 class="text-2xl font-semibold text-[#17232c]">{{ t.settings }}</h3>
          <form class="mt-5 grid gap-4" @submit.prevent="saveSettings">
            <label class="grid gap-2 text-sm text-[#40535d]">
              {{ t.autoLockMinutes }}
              <input
                v-model.number="settings.auto_lock_minutes"
                class="w-36 border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]"
                min="1"
                type="number"
              />
            </label>
            <button class="w-fit border border-[#159c8b] bg-[#dff5f1] px-5 py-3 text-sm font-semibold text-[#0d5f57] disabled:opacity-60" type="submit" :disabled="busy">
              {{ t.saveSettings }}
            </button>
          </form>
        </div>

        <p v-if="error" class="mt-5 border border-[#d98b92] bg-[#fff1f2] p-4 text-sm text-[#8e333b]">
          {{ error }}
        </p>
        <p
          v-if="notice"
          class="mt-5 border p-4 text-sm"
          :class="isLockedView ? 'border-[#d98b92] bg-[#fff1f2] text-[#8e333b]' : 'border-[#9ed2c9] bg-[#effaf8] text-[#0d5f57]'"
        >
          {{ notice }}
        </p>
      </section>
    </section>
  </main>
</template>
