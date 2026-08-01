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

type LockState = {
  initialized: boolean;
  locked: boolean;
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

type ViewMode = 'accounts' | 'generator' | 'settings';

const emptyAccount = (): AccountInput => ({
  title: '',
  username: '',
  password: '',
  url: '',
  note: '',
});

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
      eyebrow: 'First run',
      title: 'Create the local vault',
      body: 'Choose a Master Password.',
      action: 'Initialize vault',
      status: 'Uninitialized',
    };
  }

  if (state.value.locked) {
    return {
      eyebrow: 'Locked',
      title: 'Vault is sealed',
      body: 'Enter your Master Password.',
      action: 'Unlock',
      status: 'Locked',
    };
  }

  return {
    eyebrow: 'Unlocked',
    title: 'Vault workspace',
    body: `${accounts.value.length} account${accounts.value.length === 1 ? '' : 's'}`,
    action: 'Open workspace',
    status: 'Unlocked',
  };
});

const selectedSummary = computed(() => accounts.value.find((account) => account.id === editingId.value));

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

async function submitMasterPassword() {
  clearMessages();

  if (!masterPassword.value) {
    error.value = 'Master Password is required.';
    return;
  }

  busy.value = true;
  try {
    if (!state.value.initialized) {
      await InitializeVault({ master_password: masterPassword.value });
      notice.value = 'Vault initialized.';
    } else {
      await Unlock({ master_password: masterPassword.value });
      notice.value = 'Vault unlocked.';
    }
    activeView.value = 'accounts';
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
    notice.value = 'Vault locked.';
    await refreshState();
  } catch (err) {
    showError(err);
  } finally {
    busy.value = false;
  }
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
    notice.value = editingId.value ? 'Account updated.' : 'Account created.';
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
  if (!window.confirm(`Delete "${accountForm.title || 'this account'}"?`)) return;

  clearMessages();
  busy.value = true;
  try {
    await DeleteAccount(editingId.value);
    notice.value = 'Account deleted.';
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
    notice.value = 'Password generated.';
  } catch (err) {
    showError(err);
  }
}

async function generateStandalonePassword() {
  queueActivity();
  clearMessages();
  try {
    generatedPassword.value = await GeneratePassword(passwordOptions);
    notice.value = 'Password generated.';
  } catch (err) {
    showError(err);
  }
}

function useGeneratedPassword() {
  queueActivity();
  if (!generatedPassword.value) return;
  accountForm.password = generatedPassword.value;
  activeView.value = 'accounts';
  revealPassword.value = true;
  notice.value = 'Password added to the form.';
}

async function copySelectedPassword() {
  queueActivity();
  if (!editingId.value) return;

  clearMessages();
  try {
    await CopyPassword(editingId.value);
    copiedAccountId.value = editingId.value;
    notice.value = 'Password copied.';
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
    notice.value = 'Settings saved.';
  } catch (err) {
    showError(err);
  } finally {
    busy.value = false;
  }
}

function resetForm() {
  editingId.value = '';
  selected.value = null;
  revealPassword.value = false;
  Object.assign(accountForm, emptyAccount());
}

function clearMessages() {
  error.value = '';
  notice.value = '';
}

function showError(err: unknown) {
  error.value = err instanceof Error ? err.message : String(err);
}

function installRuntimeListeners() {
  removeAutoLockedListener = EventsOn('vault:auto-locked', async () => {
    notice.value = 'Vault locked after inactivity.';
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
              <p class="text-sm uppercase tracking-[0.22em] text-[#6c808c]">KEy VAult</p>
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
              Accounts
            </button>
            <button
              class="block w-full px-4 py-3 text-left"
              :class="activeView === 'generator' ? 'border-l-2 border-[#159c8b] bg-white text-[#17232c] shadow-sm' : ''"
              type="button"
              :disabled="state.locked"
              @click="activeView = 'generator'"
            >
              Password generator
            </button>
            <button
              class="block w-full px-4 py-3 text-left"
              :class="activeView === 'settings' ? 'border-l-2 border-[#159c8b] bg-white text-[#17232c] shadow-sm' : ''"
              type="button"
              :disabled="state.locked"
              @click="activeView = 'settings'"
            >
              Settings
            </button>
          </nav>
        </div>

        <div class="mt-8 grid gap-3 text-xs leading-5 text-[#6c808c]">
          <div class="border border-[#cfdee7] bg-white p-3">
            <p class="font-semibold text-[#17232c]">{{ currentMode.status }}</p>
            <p>{{ currentMode.body }}</p>
          </div>
          <p>Local first. Zero knowledge. Portable by default.</p>
        </div>
      </aside>

      <section class="w-full border border-[#cfdee7] bg-white p-5 shadow-2xl shadow-[#9fb8c8]/25 sm:p-6">
        <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
          <div>
            <p class="text-sm uppercase tracking-[0.22em] text-[#0d7f73]">{{ currentMode.eyebrow }}</p>
            <h2 class="mt-2 text-3xl font-semibold text-[#17232c] sm:text-4xl">{{ currentMode.title }}</h2>
          </div>
          <button
            v-if="state.initialized && !state.locked"
            class="border border-[#159c8b] bg-[#dff5f1] px-4 py-2 text-sm font-semibold text-[#0d5f57]"
            type="button"
            @click="lockVault"
          >
            Lock vault
          </button>
        </div>

        <div v-if="loading" class="border border-[#cfdee7] bg-[#f7fbfd] p-5 text-[#536872]">
          Checking vault state...
        </div>

        <div v-else-if="state.locked || !state.initialized" class="grid gap-5 lg:grid-cols-[1fr_240px]">
          <div class="relative overflow-hidden border border-[#cfdee7] bg-[#f7fbfd] p-5">
            <div class="absolute left-0 top-0 h-full w-1 bg-[#159c8b]"></div>
            <p class="max-w-2xl pl-3 text-base leading-7 text-[#40535d]">{{ currentMode.body }}</p>
            <form class="mt-6 grid max-w-md gap-3 pl-3" @submit.prevent="submitMasterPassword">
              <label class="text-sm font-semibold text-[#40535d]" for="master-password">Master Password</label>
              <input
                id="master-password"
                v-model="masterPassword"
                class="border border-[#b7cbd6] bg-white px-4 py-3 text-[#17232c] outline-none transition focus:border-[#159c8b]"
                type="password"
                autocomplete="current-password"
              />
              <button
                class="w-fit border border-[#159c8b] bg-[#dff5f1] px-5 py-3 text-sm font-semibold text-[#0d5f57] transition hover:bg-[#c9eee8] disabled:cursor-not-allowed disabled:opacity-60"
                type="submit"
                :disabled="busy"
              >
                {{ busy ? 'Working...' : currentMode.action }}
              </button>
            </form>
          </div>

          <div class="grid gap-3 text-sm">
            <div class="border border-[#cfdee7] bg-[#f7fbfd] p-4">
              <p class="text-[#6c808c]">Config</p>
              <p class="mt-1 text-[#17232c]">{{ state.initialized ? 'Detected' : 'Missing' }}</p>
            </div>
            <div class="border border-[#cfdee7] bg-[#f7fbfd] p-4">
              <p class="text-[#6c808c]">Session</p>
              <p class="mt-1 text-[#17232c]">{{ state.locked ? 'No key in memory' : 'Unlocked' }}</p>
            </div>
            <div class="border border-[#cfdee7] bg-[#f7fbfd] p-4">
              <p class="text-[#6c808c]">Auto-lock</p>
              <p class="mt-1 text-[#17232c]">{{ settings.auto_lock_minutes }} minutes</p>
            </div>
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
                  placeholder="Search accounts"
                  type="search"
                  @input="loadAccounts"
                />
                <button class="border border-[#b7cbd6] bg-white px-3 py-2 text-sm text-[#40535d]" type="button" @click="resetForm">
                  New
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
                <p class="mt-1 truncate text-sm text-[#6c808c]">{{ account.username || account.url || 'No username' }}</p>
              </button>
              <div v-if="accounts.length === 0" class="p-5 text-sm text-[#6c808c]">
                No accounts yet.
              </div>
            </div>
          </section>

          <section class="border border-[#cfdee7] bg-[#f7fbfd] p-5">
            <div class="mb-5 flex flex-wrap items-center justify-between gap-3">
              <div>
                <p class="text-sm uppercase tracking-[0.18em] text-[#0d7f73]">{{ editingId ? 'Edit account' : 'New account' }}</p>
                <h3 class="mt-2 text-2xl font-semibold text-[#17232c]">{{ accountForm.title || 'Untitled entry' }}</h3>
                <p v-if="selectedSummary" class="mt-1 text-sm text-[#6c808c]">Updated {{ new Date(selectedSummary.updated_at * 1000).toLocaleString() }}</p>
              </div>
            </div>

            <form class="grid gap-4" @submit.prevent="saveAccount">
              <div class="grid gap-4 md:grid-cols-2">
                <label class="grid gap-2 text-sm text-[#40535d]">
                  Title
                  <input v-model="accountForm.title" class="border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]" />
                </label>
                <label class="grid gap-2 text-sm text-[#40535d]">
                  Username
                  <input v-model="accountForm.username" class="border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]" />
                </label>
              </div>

              <label class="grid gap-2 text-sm text-[#40535d]">
                Password
                <div class="grid gap-2 sm:grid-cols-[1fr_auto_auto_auto]">
                  <input v-model="accountForm.password" class="border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]" :type="revealPassword ? 'text' : 'password'" />
                  <button class="border border-[#b7cbd6] bg-white px-3 py-2 text-sm text-[#40535d]" type="button" @click="revealPassword = !revealPassword">
                    {{ revealPassword ? 'Hide' : 'Show' }}
                  </button>
                  <button class="border border-[#b7cbd6] bg-white px-3 py-2 text-sm text-[#40535d]" type="button" @click="generateIntoForm">
                    Generate
                  </button>
                  <button class="border border-[#b7cbd6] bg-white px-3 py-2 text-sm text-[#40535d] disabled:opacity-50" type="button" :disabled="!editingId" @click="copySelectedPassword">
                    {{ copiedAccountId === editingId ? 'Copied' : 'Copy' }}
                  </button>
                </div>
              </label>

              <label class="grid gap-2 text-sm text-[#40535d]">
                URL
                <input v-model="accountForm.url" class="border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]" />
              </label>

              <label class="grid gap-2 text-sm text-[#40535d]">
                Note
                <textarea v-model="accountForm.note" class="min-h-24 border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]"></textarea>
              </label>

              <div class="flex flex-wrap gap-3">
                <button class="border border-[#159c8b] bg-[#dff5f1] px-5 py-3 text-sm font-semibold text-[#0d5f57] disabled:opacity-60" type="submit" :disabled="busy">
                  {{ editingId ? 'Save changes' : 'Create account' }}
                </button>
                <button class="border border-[#b7cbd6] bg-white px-5 py-3 text-sm font-semibold text-[#40535d]" type="button" @click="resetForm">
                  Clear
                </button>
                <button class="border border-[#d98b92] bg-[#fff1f2] px-5 py-3 text-sm font-semibold text-[#8e333b] disabled:opacity-50" type="button" :disabled="!editingId || busy" @click="removeAccount">
                  Delete
                </button>
              </div>
            </form>
          </section>
        </div>

        <div v-else-if="activeView === 'generator'" class="grid gap-5 xl:grid-cols-[360px_1fr]">
          <section class="border border-[#cfdee7] bg-[#f7fbfd] p-5">
            <h3 class="text-2xl font-semibold text-[#17232c]">Password generator</h3>
            <div class="mt-5 grid gap-5">
              <label class="grid gap-2 text-sm text-[#40535d]">
                Length: {{ passwordOptions.length }}
                <input v-model.number="passwordOptions.length" class="accent-[#159c8b]" max="128" min="8" type="range" />
              </label>
              <label class="flex items-center gap-3 text-sm text-[#40535d]">
                <input v-model="passwordOptions.include_uppercase" class="accent-[#159c8b]" type="checkbox" />
                Uppercase
              </label>
              <label class="flex items-center gap-3 text-sm text-[#40535d]">
                <input v-model="passwordOptions.include_lowercase" class="accent-[#159c8b]" type="checkbox" />
                Lowercase
              </label>
              <label class="flex items-center gap-3 text-sm text-[#40535d]">
                <input v-model="passwordOptions.include_numbers" class="accent-[#159c8b]" type="checkbox" />
                Numbers
              </label>
              <label class="flex items-center gap-3 text-sm text-[#40535d]">
                <input v-model="passwordOptions.include_symbols" class="accent-[#159c8b]" type="checkbox" />
                Symbols
              </label>
            </div>
          </section>

          <section class="border border-[#cfdee7] bg-[#f7fbfd] p-5">
            <div class="border border-[#b7cbd6] bg-white p-4 font-mono text-sm text-[#17232c] break-all">
              {{ generatedPassword || 'Generated password appears here' }}
            </div>
            <div class="mt-4 flex flex-wrap gap-3">
              <button class="border border-[#159c8b] bg-[#dff5f1] px-5 py-3 text-sm font-semibold text-[#0d5f57]" type="button" @click="generateStandalonePassword">
                Generate password
              </button>
              <button class="border border-[#b7cbd6] bg-white px-5 py-3 text-sm font-semibold text-[#40535d] disabled:opacity-50" type="button" :disabled="!generatedPassword" @click="useGeneratedPassword">
                Use in account
              </button>
            </div>
          </section>
        </div>

        <div v-else class="max-w-xl border border-[#cfdee7] bg-[#f7fbfd] p-5">
          <h3 class="text-2xl font-semibold text-[#17232c]">Settings</h3>
          <form class="mt-5 grid gap-4" @submit.prevent="saveSettings">
            <label class="grid gap-2 text-sm text-[#40535d]">
              Auto-lock minutes
              <input
                v-model.number="settings.auto_lock_minutes"
                class="w-36 border border-[#b7cbd6] bg-white px-3 py-2 text-[#17232c] outline-none focus:border-[#159c8b]"
                min="1"
                type="number"
              />
            </label>
            <button class="w-fit border border-[#159c8b] bg-[#dff5f1] px-5 py-3 text-sm font-semibold text-[#0d5f57] disabled:opacity-60" type="submit" :disabled="busy">
              Save settings
            </button>
          </form>
        </div>

        <p v-if="error" class="mt-5 border border-[#d98b92] bg-[#fff1f2] p-4 text-sm text-[#8e333b]">
          {{ error }}
        </p>
        <p v-if="notice" class="mt-5 border border-[#9ed2c9] bg-[#effaf8] p-4 text-sm text-[#0d5f57]">
          {{ notice }}
        </p>
      </section>
    </section>
  </main>
</template>
