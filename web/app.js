const state = {
  accounts: [],
  operation: null,
  accountID: null,
  activity: JSON.parse(sessionStorage.getItem("bankapp-activity") || "[]"),
};

const elements = {
  accountList: document.querySelector("#account-list"),
  accountCount: document.querySelector("#account-count"),
  totalBalance: document.querySelector("#total-balance"),
  largestBalance: document.querySelector("#largest-balance"),
  lastUpdated: document.querySelector("#last-updated"),
  serviceStatus: document.querySelector("#service-status"),
  statusDot: document.querySelector("#status-dot"),
  connectionLabel: document.querySelector("#connection-label"),
  createDialog: document.querySelector("#create-account-dialog"),
  createForm: document.querySelector("#create-account-form"),
  createSubmit: document.querySelector("#create-account-submit"),
  ownerInput: document.querySelector("#account-owner"),
  moneyDialog: document.querySelector("#money-dialog"),
  moneyForm: document.querySelector("#money-form"),
  moneyTitle: document.querySelector("#money-title"),
  moneyEyebrow: document.querySelector("#money-eyebrow"),
  moneySubmit: document.querySelector("#money-submit"),
  moneyAmount: document.querySelector("#money-amount"),
  selectedAccount: document.querySelector("#selected-account"),
  transferForm: document.querySelector("#transfer-form"),
  transferFrom: document.querySelector("#transfer-from"),
  transferTo: document.querySelector("#transfer-to"),
  transferAmount: document.querySelector("#transfer-amount"),
  transferHint: document.querySelector("#transfer-hint"),
  activityList: document.querySelector("#activity-list"),
  toastRegion: document.querySelector("#toast-region"),
};

const currency = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
});

function formatMoney(cents) {
  return currency.format(cents / 100);
}

function amountToCents(value) {
  const amount = Number(value);
  if (!Number.isFinite(amount) || amount <= 0) {
    throw new Error("Enter an amount greater than zero.");
  }
  return Math.round((amount + Number.EPSILON) * 100);
}

function escapeHTML(value) {
  return String(value).replace(/[&<>'"]/g, (character) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    "'": "&#39;",
    '"': "&quot;",
  })[character]);
}

function initials(owner) {
  return owner
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() || "")
    .join("");
}

async function api(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
  });

  const payload = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error(payload?.error?.message || `Request failed with status ${response.status}.`);
  }
  return payload;
}

function setConnection(online) {
  elements.statusDot.classList.toggle("online", online);
  elements.statusDot.classList.toggle("offline", !online);
  elements.connectionLabel.textContent = online ? "Connected" : "Unavailable";
  elements.serviceStatus.textContent = online ? "Online" : "Offline";
}

async function loadAccounts({ quiet = false } = {}) {
  if (!quiet) {
    elements.accountList.innerHTML = `
      <div class="empty-state loading-state">
        <span class="loader"></span>
        <p>Loading your accounts…</p>
      </div>`;
  }

  try {
    state.accounts = await api("/api/accounts");
    setConnection(true);
    renderDashboard();
  } catch (error) {
    setConnection(false);
    elements.accountList.innerHTML = `
      <div class="empty-state">
        <p>We couldn't load your accounts. Check that BankApp and PostgreSQL are running.</p>
        <button class="button button-dark" type="button" data-retry>Try again</button>
      </div>`;
    elements.accountList.querySelector("[data-retry]")?.addEventListener("click", () => loadAccounts());
    showToast("Connection problem", error.message, true);
  }
}

function renderDashboard() {
  const total = state.accounts.reduce((sum, account) => sum + account.balance, 0);
  const largest = state.accounts.reduce((max, account) => Math.max(max, account.balance), 0);

  elements.totalBalance.textContent = formatMoney(total);
  elements.largestBalance.textContent = formatMoney(largest);
  elements.accountCount.textContent = state.accounts.length;
  elements.lastUpdated.textContent = `Updated ${new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;

  renderAccounts();
  renderTransferOptions();
}

function renderAccounts() {
  if (state.accounts.length === 0) {
    elements.accountList.innerHTML = `
      <div class="empty-state">
        <p>No accounts yet. Create your first account and start moving money.</p>
        <button class="button button-dark" type="button" data-create-first>Create account</button>
      </div>`;
    elements.accountList.querySelector("[data-create-first]")?.addEventListener("click", openCreateDialog);
    return;
  }

  elements.accountList.innerHTML = state.accounts.map((account) => `
    <article class="account-row">
      <span class="account-avatar" aria-hidden="true">${escapeHTML(initials(account.owner))}</span>
      <div class="account-details">
        <strong>${escapeHTML(account.owner)}</strong>
        <span>Account ···· ${String(account.id).padStart(4, "0")}</span>
      </div>
      <strong class="account-balance">${formatMoney(account.balance)}</strong>
      <div class="account-actions">
        <button class="mini-button" type="button" data-operation="deposit" data-account-id="${account.id}" data-short-label="+" aria-label="Deposit into ${escapeHTML(account.owner)}">+ Add</button>
        <button class="mini-button" type="button" data-operation="withdraw" data-account-id="${account.id}" data-short-label="−" aria-label="Withdraw from ${escapeHTML(account.owner)}">− Take</button>
      </div>
    </article>
  `).join("");
}

function renderTransferOptions() {
  const previousFrom = Number(elements.transferFrom.value);
  const previousTo = Number(elements.transferTo.value);
  const options = state.accounts.map((account) => (
    `<option value="${account.id}">${escapeHTML(account.owner)} · ${formatMoney(account.balance)}</option>`
  )).join("");

  elements.transferFrom.innerHTML = options || '<option value="">No accounts available</option>';
  elements.transferTo.innerHTML = options || '<option value="">No accounts available</option>';

  if (state.accounts.some((account) => account.id === previousFrom)) {
    elements.transferFrom.value = String(previousFrom);
  }
  if (state.accounts.some((account) => account.id === previousTo)) {
    elements.transferTo.value = String(previousTo);
  } else if (state.accounts.length > 1) {
    elements.transferTo.selectedIndex = 1;
  }

  const submit = elements.transferForm.querySelector('button[type="submit"]');
  submit.disabled = state.accounts.length < 2;
  elements.transferHint.textContent = state.accounts.length < 2
    ? "Create at least two accounts to make a transfer."
    : "Transfers update both accounts together.";
}

function openCreateDialog() {
  elements.createForm.reset();
  elements.createDialog.showModal();
  requestAnimationFrame(() => elements.ownerInput.focus());
}

function openMoneyDialog(operation, accountID) {
  const account = state.accounts.find((item) => item.id === accountID);
  if (!account) return;

  state.operation = operation;
  state.accountID = accountID;
  elements.moneyForm.reset();
  elements.moneyTitle.textContent = operation === "deposit" ? "Add money" : "Withdraw money";
  elements.moneyEyebrow.textContent = operation === "deposit" ? "Grow the balance" : "Take from balance";
  elements.moneySubmit.textContent = operation === "deposit" ? "Add money" : "Withdraw";
  elements.selectedAccount.innerHTML = `
    <strong>${escapeHTML(account.owner)}</strong>
    <span>${formatMoney(account.balance)} available</span>`;
  elements.moneyDialog.showModal();
  requestAnimationFrame(() => elements.moneyAmount.focus());
}

function setBusy(button, busy, busyLabel = "Working…") {
  if (busy) {
    button.dataset.originalLabel = button.textContent;
    button.textContent = busyLabel;
    button.disabled = true;
  } else {
    button.textContent = button.dataset.originalLabel || button.textContent;
    button.disabled = false;
  }
}

function recordActivity(type, title, detail) {
  state.activity.unshift({
    id: crypto.randomUUID(),
    type,
    title,
    detail,
    time: new Date().toISOString(),
  });
  state.activity = state.activity.slice(0, 8);
  sessionStorage.setItem("bankapp-activity", JSON.stringify(state.activity));
  renderActivity();
}

function renderActivity() {
  if (state.activity.length === 0) {
    elements.activityList.innerHTML = `
      <div class="empty-activity">
        <span>◎</span>
        <p>Your deposits, withdrawals, and transfers will appear here.</p>
      </div>`;
    return;
  }

  const icons = { deposit: "+", withdraw: "−", transfer: "↗", account: "○" };
  elements.activityList.innerHTML = state.activity.map((item) => `
    <article class="activity-item">
      <span class="activity-icon" aria-hidden="true">${icons[item.type] || "·"}</span>
      <div class="activity-copy">
        <strong>${escapeHTML(item.title)}</strong>
        <span>${escapeHTML(item.detail)}</span>
      </div>
      <time class="activity-time" datetime="${item.time}">${new Date(item.time).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</time>
    </article>
  `).join("");
}

function showToast(title, message, isError = false) {
  const toast = document.createElement("div");
  toast.className = `toast${isError ? " error" : ""}`;
  toast.innerHTML = `<div><strong>${escapeHTML(title)}</strong><span>${escapeHTML(message)}</span></div>`;
  elements.toastRegion.append(toast);
  setTimeout(() => toast.remove(), 4200);
}

document.querySelector("#today-label").textContent = new Intl.DateTimeFormat("en-US", {
  weekday: "long",
  month: "long",
  day: "numeric",
}).format(new Date());

document.querySelector("#open-create-account").addEventListener("click", openCreateDialog);
document.querySelector("#refresh-accounts").addEventListener("click", () => loadAccounts({ quiet: true }));

elements.accountList.addEventListener("click", (event) => {
  const button = event.target.closest("[data-operation]");
  if (!button) return;
  openMoneyDialog(button.dataset.operation, Number(button.dataset.accountId));
});

elements.createForm.addEventListener("submit", async (event) => {
  if (event.submitter?.value === "cancel") return;
  event.preventDefault();

  const owner = elements.ownerInput.value.trim();
  if (!owner) return;

  setBusy(elements.createSubmit, true, "Creating…");
  try {
    const account = await api("/api/accounts", {
      method: "POST",
      body: JSON.stringify({ owner }),
    });
    elements.createDialog.close();
    recordActivity("account", "Account created", `${account.owner} is ready to use`);
    showToast("Account created", `${account.owner}'s account is ready.`);
    await loadAccounts({ quiet: true });
  } catch (error) {
    showToast("Could not create account", error.message, true);
  } finally {
    setBusy(elements.createSubmit, false);
  }
});

elements.moneyForm.addEventListener("submit", async (event) => {
  if (event.submitter?.value === "cancel") return;
  event.preventDefault();

  try {
    const amount = amountToCents(elements.moneyAmount.value);
    const account = state.accounts.find((item) => item.id === state.accountID);
    setBusy(elements.moneySubmit, true);
    await api(`/api/accounts/${state.accountID}/${state.operation}`, {
      method: "POST",
      body: JSON.stringify({ amount }),
    });
    elements.moneyDialog.close();
    const pastTense = state.operation === "deposit" ? "Deposited" : "Withdrew";
    recordActivity(state.operation, `${pastTense} ${formatMoney(amount)}`, account.owner);
    showToast("Balance updated", `${account.owner}'s balance was updated.`);
    await loadAccounts({ quiet: true });
  } catch (error) {
    showToast("Could not update balance", error.message, true);
  } finally {
    setBusy(elements.moneySubmit, false);
  }
});

elements.transferForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  try {
    const fromID = Number(elements.transferFrom.value);
    const toID = Number(elements.transferTo.value);
    const amount = amountToCents(elements.transferAmount.value);
    if (fromID === toID) throw new Error("Choose two different accounts.");

    const from = state.accounts.find((account) => account.id === fromID);
    const to = state.accounts.find((account) => account.id === toID);
    const submit = elements.transferForm.querySelector('button[type="submit"]');
    setBusy(submit, true, "Sending…");
    await api("/api/transfers", {
      method: "POST",
      body: JSON.stringify({
        from_account_id: fromID,
        to_account_id: toID,
        amount,
      }),
    });
    elements.transferAmount.value = "";
    recordActivity("transfer", `Transferred ${formatMoney(amount)}`, `${from.owner} → ${to.owner}`);
    showToast("Transfer complete", `${formatMoney(amount)} moved to ${to.owner}.`);
    await loadAccounts({ quiet: true });
  } catch (error) {
    showToast("Transfer failed", error.message, true);
  } finally {
    const submit = elements.transferForm.querySelector('button[type="submit"]');
    setBusy(submit, false);
    submit.disabled = state.accounts.length < 2;
  }
});

document.querySelector("#swap-accounts").addEventListener("click", () => {
  const from = elements.transferFrom.value;
  elements.transferFrom.value = elements.transferTo.value;
  elements.transferTo.value = from;
});

document.querySelector("#clear-activity").addEventListener("click", () => {
  state.activity = [];
  sessionStorage.removeItem("bankapp-activity");
  renderActivity();
});

for (const dialog of document.querySelectorAll("dialog")) {
  dialog.addEventListener("click", (event) => {
    if (event.target === dialog) dialog.close();
  });
}

renderActivity();
loadAccounts();
