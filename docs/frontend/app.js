const API = "/api";
const TOKEN_KEY = "cms_token";

function el(tag, attrs = {}, children = []) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === "class") node.className = v;
    else if (k === "text") node.textContent = v;
    else if (k.startsWith("on") && typeof v === "function") node.addEventListener(k.slice(2).toLowerCase(), v);
    else if (v !== undefined && v !== null) node.setAttribute(k, String(v));
  }
  for (const c of Array.isArray(children) ? children : [children]) {
    if (c === null || c === undefined) continue;
    node.appendChild(typeof c === "string" ? document.createTextNode(c) : c);
  }
  return node;
}

function toast(msg) {
  const t = el("div", { class: "toast", text: msg });
  document.body.appendChild(t);
  setTimeout(() => t.remove(), 3200);
}

function getToken() {
  return localStorage.getItem(TOKEN_KEY) || "";
}

function setToken(token) {
  if (!token) localStorage.removeItem(TOKEN_KEY);
  else localStorage.setItem(TOKEN_KEY, token);
}

async function apiFetch(path, opts = {}) {
  const headers = new Headers(opts.headers || {});
  const token = getToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);
  const res = await fetch(`${API}${path}`, { ...opts, headers });
  if (res.status === 401) {
    setToken("");
    appState.route = "login";
    render();
    throw new Error("unauthorized");
  }
  const ct = res.headers.get("content-type") || "";
  const body = ct.includes("application/json") ? await res.json() : await res.text();
  if (!res.ok) {
    const textBody = typeof body === "string" ? body.trim() : "";
    const msg = body && body.message
      ? body.message
      : textBody && !ct.includes("text/html") && !textBody.startsWith("<") && textBody.length <= 200
        ? textBody
        : `请求失败（HTTP ${res.status}）`;
    throw new Error(msg);
  }
  return body;
}

function apiUrl(path, params = {}) {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "" || Number.isNaN(v)) continue;
    q.set(k, String(v));
  }
  return `${API}${path}${q.toString() ? `?${q}` : ""}`;
}

function newRequestId() {
  if (globalThis.crypto && typeof globalThis.crypto.randomUUID === "function") {
    return globalThis.crypto.randomUUID();
  }
  return `score-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

function downloadWithAuth(path, params = {}, filename = "download.xlsx") {
  const token = getToken();
  if (!token) throw new Error("未登录");
  const xhr = new XMLHttpRequest();
  xhr.open("GET", apiUrl(path, params), true);
  xhr.responseType = "blob";
  xhr.setRequestHeader("Authorization", `Bearer ${token}`);
  xhr.onload = () => {
    if (xhr.status >= 200 && xhr.status < 300) {
      const blob = xhr.response;
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      setTimeout(() => URL.revokeObjectURL(url), 1500);
      return;
    }
    toast(`导出失败（${xhr.status}）`);
  };
  xhr.onerror = () => toast("导出失败（网络错误）");
  xhr.send();
}

function createTimerState(min = 5, sec = 0) {
  const safeMin = Math.max(0, Math.min(240, Math.floor(Number(min) || 0)));
  const safeSec = Math.max(0, Math.min(59, Math.floor(Number(sec) || 0)));
  const targetMs = (safeMin * 60 + safeSec) * 1000;
  return {
    mode: "countdown",
    running: false,
    presetMin: safeMin,
    presetSec: safeSec,
    targetMs,
    leftMs: targetMs,
    lastTick: 0,
    timeoutId: 0,
  };
}

const appState = {
  route: getToken() ? "students" : "login",
  me: null,
  studentEditor: { id: 0, studentNo: "", name: "", gender: "", phone: "", position: "", groupId: 0 },
  scoreDraft: {
    scope: "student",
    targetId: 0,
  },
  pendingScoreRequest: null,
  students: { total: 0, items: [] },
  studentsQuery: { page: 1, size: 20, keyword: "", groupId: 0 },
  groups: [],
  dimensions: [],
  scoreItems: [],
  recentScoreItems: [],
  scoreItemFilterDimensionId: 0,
  scoreTargetKeyword: "",
  rankingsType: "students",
  rankings: [],
  groupRankings: [],
  dimensionStats: { dimensionId: 0, startDate: "", endDate: "", items: [] },
  scoreEntries: { total: 0, items: [] },
  scoreEntriesQuery: { page: 1, size: 20, studentId: 0, groupId: 0, sinceDays: 30 },
  scoreEntriesStudentKeyword: "",
  rollcallPickCount: 1,
  rollcall: { roundId: "", students: [], remaining: 0 },
  groupsStudentKeyword: "",
  timers: [createTimerState(5, 0), createTimerState(5, 0)],
};

function stopTimerLoop(timerIndex) {
  const timer = appState.timers[timerIndex];
  if (!timer) return;
  if (timer.timeoutId) {
    clearTimeout(timer.timeoutId);
    timer.timeoutId = 0;
  }
}

function syncTimerUI(timerIndex) {
  const timer = appState.timers[timerIndex];
  if (!timer) return;
  const uiIndex = timerIndex + 1;
  const clock = document.getElementById(`timer-clock-${uiIndex}`);
  if (clock) clock.textContent = fmtClock(timer.leftMs);
  const btn = document.getElementById(`timer-start-btn-${uiIndex}`);
  if (btn) btn.textContent = timer.running ? "暂停" : "开始";
}

function clearScoreDraft() {
  appState.scoreDraft = { scope: "student", targetId: 0 };
}

function pad2(n) {
  return String(n).padStart(2, "0");
}

function fmtClock(ms) {
  const s = Math.max(0, Math.floor(ms / 1000));
  const mm = Math.floor(s / 60);
  const ss = s % 60;
  return `${pad2(mm)}:${pad2(ss)}`;
}

async function loadBootstrap() {
  try {
    appState.me = await apiFetch("/auth/me");
  } catch {
    return;
  }
}

async function loadStudents() {
  await loadStudentsAll();
}

async function loadStudentsList(params = {}) {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "" || Number.isNaN(v)) continue;
    if (typeof v === "number" && v === 0 && k === "groupId") continue;
    q.set(k, String(v));
  }
  appState.students = await apiFetch(`/students${q.toString() ? `?${q}` : ""}`);
}

async function loadStudentsAll() {
  const pageSize = 200;
  let page = 1;
  let total = 0;
  const items = [];
  for (let i = 0; i < 50; i++) {
    const res = await apiFetch(`/students?page=${page}&size=${pageSize}`);
    const batch = res.items || [];
    total = Number(res.total || total || 0);
    for (const it of batch) items.push(it);
    if (batch.length < pageSize) break;
    if (total && items.length >= total) break;
    page += 1;
  }
  appState.students = { total: total || items.length, items };
}

async function loadStudentsForPickers() {
  await loadStudentsAll();
}

async function studentCreate(payload) {
  return apiFetch("/students", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

async function studentUpdate(studentId, payload) {
  return apiFetch(`/students/${studentId}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

async function studentDelete(studentId) {
  await apiFetch(`/students/${studentId}`, { method: "DELETE" });
}

async function loadGroups() {
  const res = await apiFetch("/groups");
  appState.groups = res.items || [];
}

async function loadDimensions() {
  const res = await apiFetch("/dimensions");
  appState.dimensions = res.items || [];
}

async function loadScoreItems() {
  const res = await apiFetch("/score-items");
  appState.scoreItems = res.items || [];
}

async function loadRecentScoreItems() {
  const res = await apiFetch("/score-items/recent");
  appState.recentScoreItems = res.items || [];
}

async function fetchRankings(params = {}) {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "" || Number.isNaN(v)) continue;
    q.set(k, String(v));
  }
  const res = await apiFetch(`/rankings/students${q.toString() ? `?${q}` : ""}`);
  return res.items || [];
}

async function loadRankings(params = {}) {
  appState.rankings = await fetchRankings(params);
}

async function loadDimensionStats(params) {
  appState.dimensionStats = { ...appState.dimensionStats, ...params, items: await fetchRankings(params) };
}

async function loadGroupRankings(params = {}) {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "" || Number.isNaN(v)) continue;
    q.set(k, String(v));
  }
  const res = await apiFetch(`/rankings/groups${q.toString() ? `?${q}` : ""}`);
  appState.groupRankings = res.items || [];
}

async function loadScoreEntries(params = {}) {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "" || Number.isNaN(v)) continue;
    if (typeof v === "number" && v === 0 && (k === "studentId" || k === "groupId")) continue;
    q.set(k, String(v));
  }
  const res = await apiFetch(`/score-entries${q.toString() ? `?${q}` : ""}`);
  appState.scoreEntries = { total: Number(res.total || 0), items: res.items || [] };
}

async function groupCreate(name) {
  return apiFetch("/groups", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
}

async function groupUpdate(id, name) {
  return apiFetch(`/groups/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
}

async function groupDelete(id) {
  await apiFetch(`/groups/${id}`, { method: "DELETE" });
}

async function studentAssignGroup(studentId, groupId) {
  await apiFetch(`/students/${studentId}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ groupId }),
  });
}

async function dimensionCreate(name) {
  return apiFetch("/dimensions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
}

async function dimensionUpdate(id, name) {
  return apiFetch(`/dimensions/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
}

async function dimensionDelete(id) {
  await apiFetch(`/dimensions/${id}`, { method: "DELETE" });
}

async function scoreItemCreate(payload) {
  return apiFetch("/score-items", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

async function scoreItemUpdate(id, payload) {
  return apiFetch(`/score-items/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

async function scoreItemDelete(id) {
  await apiFetch(`/score-items/${id}`, { method: "DELETE" });
}

async function scoreEntryCreate(payload) {
  await apiFetch("/score-entries", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

async function scoreEntryDelete(id) {
  await apiFetch(`/score-entries/${id}`, { method: "DELETE" });
}

async function rollcallStart(fair) {
  appState.rollcall = await apiFetch("/rollcall/start", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ fair: !!fair, count: Number(appState.rollcallPickCount || 1) }),
  });
}

async function rollcallPick() {
  const c = Math.max(1, Math.min(50, Math.floor(Number(appState.rollcallPickCount || 1))));
  appState.rollcall = await apiFetch(`/rollcall/pick?count=${c}`, { method: "POST" });
}

async function rollcallReset(roundId) {
  await apiFetch("/rollcall/reset", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ roundId: roundId || "" }),
  });
  appState.rollcall = { roundId: "", students: [], remaining: 0 };
}

function viewLogin() {
  const username = el("input", { type: "text", autocomplete: "username", placeholder: "请输入用户名", autofocus: "" });
  const password = el("input", { type: "password", autocomplete: "current-password", placeholder: "请输入密码" });
  const btn = el("button", {
    class: "btn btn-amber",
    text: "登录",
    type: "submit",
  });

  const form = el("form", {
    class: "login-form",
    onsubmit: async (event) => {
      event.preventDefault();
      if (btn.disabled) return;
      btn.disabled = true;
      btn.textContent = "正在登录...";
      try {
        const res = await fetch(`${API}/auth/login`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ username: username.value, password: password.value }),
        });
        const ct = res.headers.get("content-type") || "";
        const body = ct.includes("application/json") ? await res.json() : await res.text();
        if (!res.ok) {
          const msg = body && body.message ? body.message : `登录失败（HTTP ${res.status}）`;
          throw new Error(msg);
        }
        setToken(body.accessToken);
        await loadBootstrap();
        appState.route = "students";
        render();
      } catch (e) {
        toast(String(e.message || e));
        btn.disabled = false;
        btn.textContent = "登录";
      }
    },
  }, [
    el("div", { class: "field" }, [el("label", { text: "用户名" }), username]),
    el("div", { class: "field" }, [el("label", { text: "密码" }), password]),
    btn,
  ]);

  return el("main", { class: "login-screen" }, [
    el("section", { class: "card login-card", "aria-labelledby": "login-title" }, [
      el("div", { class: "login-heading" }, [
        el("p", { class: "login-eyebrow", text: "智慧班级" }),
        el("h1", { id: "login-title", text: "欢迎回来" }),
      ]),
      form,
    ]),
  ]);
}

function shell(title, content) {
  const navItems = [
    { key: "students", label: "学生名单" },
    { key: "score", label: "积分录入" },
    { key: "rollcall", label: "随机点名" },
    { key: "groups", label: "小组管理" },
    { key: "config", label: "维度与积分项" },
    { key: "entries", label: "积分记录" },
    { key: "rankings", label: "排行榜" },
    { key: "timer", label: "计时器" },
  ];

  async function loadForRoute(routeKey) {
    if (routeKey === "students") {
      await loadGroups();
      await loadStudentsList(appState.studentsQuery);
    }
    if (routeKey === "score") {
      await Promise.all([loadStudentsForPickers(), loadGroups(), loadDimensions(), loadScoreItems(), loadRecentScoreItems()]);
    }
    if (routeKey === "groups") {
      await Promise.all([loadStudentsForPickers(), loadGroups()]);
    }
    if (routeKey === "config") {
      await Promise.all([loadStudentsForPickers(), loadDimensions(), loadScoreItems()]);
    }
    if (routeKey === "entries") {
      await Promise.all([
        loadStudentsForPickers(),
        loadGroups(),
        loadDimensions(),
        loadScoreItems(),
        loadScoreEntries(appState.scoreEntriesQuery),
      ]);
    }
    if (routeKey === "rankings") {
      await Promise.all([loadDimensions(), loadRankings({})]);
    }
  }

  const nav = el("div", { class: "nav" }, navItems.map((it) =>
    el("button", {
      class: it.key === appState.route ? "active" : "",
      text: it.label,
      onclick: async () => {
        appState.route = it.key;
        try {
          await loadForRoute(it.key);
        } catch (e) {
          toast(String(e.message || e));
        }
        render();
      },
    })
  ));

  const logout = el("button", {
    class: "btn",
    text: "退出登录",
    onclick: () => {
      setToken("");
      appState.route = "login";
      render();
    },
  });

  return el("div", { class: "shell" }, [
    el("aside", { class: "sidebar" }, [
      el("div", { class: "brand" }, [
        el("h1", { text: "智慧班级综合管理系统" }),
        el("p", { text: appState.me?.username ? `已登录：${appState.me.username}` : "" }),
      ]),
      nav,
    ]),
    el("main", { class: "main" }, [
      el("div", { class: "topbar" }, [
        el("h2", { class: "title", text: title }),
        logout,
      ]),
      content,
    ]),
  ]);
}

function groupNameById(id) {
  const g = (appState.groups || []).find((x) => Number(x.id) === Number(id));
  return g ? g.name : "未分组";
}

function studentNameById(id) {
  const s = (appState.students.items || []).find((x) => Number(x.id) === Number(id));
  return s ? `${s.studentNo} ${s.name}` : `学生ID ${id}`;
}

function dimensionNameById(id) {
  const d = (appState.dimensions || []).find((x) => Number(x.id) === Number(id));
  return d ? d.name : `维度ID ${id}`;
}

function scoreItemById(id) {
  return (appState.scoreItems || []).find((x) => Number(x.id) === Number(id)) || null;
}

function viewStudents() {
  const q = { ...appState.studentsQuery };
  const reload = el("button", {
    class: "btn btn-amber",
    text: "刷新",
    onclick: async () => {
      try {
        await loadStudentsList(appState.studentsQuery);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });
  
  const tpl = el("a", {
    class: "btn",
    href: "/templates/students_import_template.xlsx",
    text: "下载导入模板",
    download: "students_import_template.xlsx",
  });

  const file = el("input", { type: "file", accept: ".xlsx" });
  const upload = el("button", {
    class: "btn",
    text: "Excel 导入",
    onclick: async () => {
      try {
        if (!file.files || file.files.length === 0) throw new Error("请选择 .xlsx 文件");
        const fd = new FormData();
        fd.append("file", file.files[0]);
        await apiFetch("/students/import", { method: "POST", body: fd });
        toast("导入成功");
        await loadStudentsList(appState.studentsQuery);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  const kw = el("input", { type: "text", placeholder: "搜索：姓名/学号", value: q.keyword || "" });
  const groupFilter = el("select");
  groupFilter.appendChild(el("option", { value: "0", text: "全部小组" }));
  for (const g of appState.groups || []) {
    groupFilter.appendChild(el("option", { value: String(g.id), text: g.name }));
  }
  groupFilter.value = String(q.groupId || 0);
  const size = el("input", { type: "number", min: "1", max: "200" });
  size.value = String(q.size || 20);
  const query = el("button", {
    class: "btn btn-amber",
    text: "查询",
    onclick: async () => {
      try {
        appState.studentsQuery = {
          page: 1,
          size: Math.max(1, Math.min(200, Number(size.value || 20))),
          keyword: (kw.value || "").trim(),
          groupId: Number(groupFilter.value || 0),
        };
        await loadStudentsList(appState.studentsQuery);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });
  const prev = el("button", {
    class: "btn",
    text: "上一页",
    onclick: async () => {
      try {
        const cur = appState.studentsQuery.page || 1;
        if (cur <= 1) return;
        appState.studentsQuery = { ...appState.studentsQuery, page: cur - 1 };
        await loadStudentsList(appState.studentsQuery);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });
  const next = el("button", {
    class: "btn",
    text: "下一页",
    onclick: async () => {
      try {
        const cur = appState.studentsQuery.page || 1;
        const sz = appState.studentsQuery.size || 20;
        const total = appState.students.total || 0;
        if (cur * sz >= total) return;
        appState.studentsQuery = { ...appState.studentsQuery, page: cur + 1 };
        await loadStudentsList(appState.studentsQuery);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  const ed = { ...appState.studentEditor };
  const edNo = el("input", { type: "text", placeholder: "学号", value: ed.studentNo || "" });
  const edName = el("input", { type: "text", placeholder: "姓名", value: ed.name || "" });
  const edGender = el("select", {}, [
    el("option", { value: "", text: "（可选）" }),
    el("option", { value: "男", text: "男" }),
    el("option", { value: "女", text: "女" }),
  ]);
  edGender.value = ed.gender || "";
  const edPhone = el("input", { type: "text", placeholder: "联系方式（可选）", value: ed.phone || "" });
  const edPos = el("input", { type: "text", placeholder: "班委职位（可选）", value: ed.position || "" });
  const edGroup = el("select");
  edGroup.appendChild(el("option", { value: "0", text: "未分组" }));
  for (const g of appState.groups || []) {
    edGroup.appendChild(el("option", { value: String(g.id), text: g.name }));
  }
  edGroup.value = String(ed.groupId || 0);

  const saveEditor = el("button", {
    class: "btn btn-amber",
    text: ed.id ? "保存修改" : "新增学生",
    onclick: async () => {
      try {
        const payload = {
          studentNo: (edNo.value || "").trim(),
          name: (edName.value || "").trim(),
          gender: (edGender.value || "").trim(),
          phone: (edPhone.value || "").trim(),
          position: (edPos.value || "").trim(),
          groupId: Number(edGroup.value || 0),
        };
        if (!payload.studentNo || !payload.name) throw new Error("请填写学号与姓名");

        if (!ed.id) {
          const created = await studentCreate(payload);
          const gid = Number(payload.groupId || 0);
          if (gid) await studentUpdate(created.id, { groupId: gid });
          toast("已创建");
        } else {
          await studentUpdate(ed.id, payload);
          toast("已保存");
        }
        appState.studentEditor = { id: 0, studentNo: "", name: "", gender: "", phone: "", position: "", groupId: 0 };
        await Promise.all([loadStudentsList(appState.studentsQuery), loadGroups()]);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });
  const cancelEditor = el("button", {
    class: "btn",
    text: "取消",
    onclick: () => {
      appState.studentEditor = { id: 0, studentNo: "", name: "", gender: "", phone: "", position: "", groupId: 0 };
      render();
    },
  });
  const delEditor = ed.id
    ? el("button", {
        class: "btn btn-danger",
        text: "删除",
        onclick: async () => {
          try {
            const ok = window.confirm(`确认删除学生「${ed.name || edName.value || ""}」？`);
            if (!ok) return;
            await studentDelete(ed.id);
            toast("已删除");
            appState.studentEditor = { id: 0, studentNo: "", name: "", gender: "", phone: "", position: "", groupId: 0 };
            await loadStudentsList(appState.studentsQuery);
            render();
          } catch (e) {
            toast(String(e.message || e));
          }
        },
      })
    : null;

  const list = el("div", { class: "list" }, (appState.students.items || []).map((s) => {
    const gender = (s.gender || "").trim() || "-";
    const group = groupNameById(s.groupId);
    const edit = el("button", {
      class: "btn btn-small",
      text: "编辑",
      onclick: () => {
        appState.studentEditor = {
          id: s.id,
          studentNo: s.studentNo || "",
          name: s.name || "",
          gender: s.gender || "",
          phone: s.phone || "",
          position: s.position || "",
          groupId: Number(s.groupId || 0),
        };
        render();
      },
    });

    return el("div", { class: "student-item" }, [
      el("div", { class: "student-name" }, [
        el("div", { class: "combo" }, [
          el("strong", { text: String(s.studentNo || "") }),
          " ",
          el("strong", { text: String(s.name || "") }),
        ]),
        el("div", { class: "meta", text: `${gender} · ${group}` }),
      ]),
      el("div", { class: "row" }, [
        edit,
      ]),
    ]);
  }));

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "学生名单" }),
      el("div", { class: "row" }, [
        el("span", { class: "pill" }, [el("span", { text: `共 ${appState.students.total || 0} 人` })]),
        el("span", { class: "pill" }, [el("span", { text: `第 ${appState.studentsQuery.page || 1} 页` })]),
        reload,
        tpl,
        file,
        upload,
      ]),
      el("div", { class: "sep" }),
      el("div", { class: "row" }, [
        el("div", { class: "field" }, [el("label", { text: "关键词" }), kw]),
        el("div", { class: "field" }, [el("label", { text: "小组" }), groupFilter]),
        el("div", { class: "field" }, [el("label", { text: "每页条数" }), size]),
        query,
        prev,
        next,
      ]),
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: ed.id ? "编辑学生" : "新增学生" }),
      el("div", { class: "row" }, [edNo, edName, edGender, edPhone, edPos, edGroup, saveEditor, cancelEditor, delEditor]),
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: "列表" }),
      list,
    ]),
  ]);

  return shell("学生名单", content);
}

function viewScore() {
  const scopeSel = el("select", {}, [
    el("option", { value: "student", text: "单人" }),
    el("option", { value: "group", text: "小组" }),
    el("option", { value: "class", text: "全班" }),
  ]);
  const targetSel = el("select");
  const targetKw = el("input", { type: "text", placeholder: "搜索（学号/姓名）" });
  targetKw.value = String(appState.scoreTargetKeyword || "");
  const remark = el("input", { type: "text", placeholder: "备注（可选）" });


  const filterDim = el("select");
  filterDim.appendChild(el("option", { value: "0", text: "全部维度" }));
  for (const d of appState.dimensions || []) {
    filterDim.appendChild(el("option", { value: String(d.id), text: d.name }));
  }
  filterDim.value = String(appState.scoreItemFilterDimensionId || 0);

  function tokenizeKw(v) {
    const s = String(v || "").trim().toLowerCase();
    if (!s) return [];
    return s.split(/\s+/).filter(Boolean);
  }

  function matchAllTokens(haystack, tokens) {
    if (!tokens.length) return true;
    const h = String(haystack || "").toLowerCase();
    return tokens.every((t) => h.includes(t));
  }

  function refreshTargets() {
    targetSel.disabled = false;
    const prev = String(targetSel.value || "");
    targetSel.innerHTML = "";

    targetKw.style.display = scopeSel.value === "class" ? "none" : "";
    targetKw.placeholder = scopeSel.value === "group" ? "搜索（小组名）" : "搜索（学号/姓名）";

    const kwTokens = tokenizeKw(targetKw.value);
    if (scopeSel.value === "student") {
      for (const s of appState.students.items || []) {
        const txt = `${s.studentNo || ""} ${s.name || ""}`.trim();
        if (!matchAllTokens(txt, kwTokens)) continue;
        targetSel.appendChild(el("option", { value: String(s.id), text: txt }));
      }
    } else if (scopeSel.value === "group") {
      for (const g of appState.groups || []) {
        if (!matchAllTokens(g.name || "", kwTokens)) continue;
        targetSel.appendChild(el("option", { value: String(g.id), text: g.name }));
      }
    } else {
      targetSel.appendChild(el("option", { value: "0", text: "全班" }));
    }

    if (scopeSel.value !== "class" && !targetSel.options.length) {
      targetSel.appendChild(el("option", { value: "", text: "无匹配对象" }));
      targetSel.value = "";
      targetSel.disabled = true;
      return;
    }

    if (appState.scoreDraft && appState.scoreDraft.scope === scopeSel.value) {
      const v = String(appState.scoreDraft.targetId || 0);
      if ([...targetSel.options].some((o) => o.value === v)) {
        targetSel.value = v;
      }
    } else if (prev && [...targetSel.options].some((o) => o.value === prev)) {
      targetSel.value = prev;
    }
  }

  scopeSel.addEventListener("change", refreshTargets);
  scopeSel.value = appState.scoreDraft?.scope || "student";
  refreshTargets();

  const currentScore = el("div", { class: "pill" }, [el("span", { text: "当前积分 -" })]);
  function syncCurrentScore() {
    if (scopeSel.value !== "student") {
      currentScore.style.display = "none";
      return;
    }
    currentScore.style.display = "";
    if (targetSel.disabled || !String(targetSel.value || "")) {
      currentScore.textContent = "当前积分 -";
      return;
    }
    const sid = Number(targetSel.value || 0);
    const st = (appState.students.items || []).find((x) => Number(x.id) === sid) || null;
    const sc = st ? Number(st.totalScore || 0) : 0;
    currentScore.textContent = `当前积分 ${sc}`;
  }
  targetSel.addEventListener("change", syncCurrentScore);
  scopeSel.addEventListener("change", syncCurrentScore);
  targetKw.addEventListener("input", () => {
    appState.scoreTargetKeyword = targetKw.value || "";
    refreshTargets();
    syncCurrentScore();
  });
  syncCurrentScore();

  function itemBtn(it) {
    const score = Number(it.score || 0);
    const cls = score >= 0 ? "btn btn-green" : "btn btn-red";
    const txt = `${it.name} ${score >= 0 ? "+" : ""}${score}`;
    const button = el("button", {
      class: cls,
      text: txt,
      onclick: async () => {
        button.disabled = true;
        try {
          if (scopeSel.value !== "class" && targetSel.disabled) {
            throw new Error("请选择对象");
          }
          const payload = {
            scope: scopeSel.value,
            targetId: scopeSel.value === "class" ? 0 : Number(targetSel.value || 0),
            scoreItemId: Number(it.id),
            remark: remark.value || "",
          };
          const signature = JSON.stringify(payload);
          if (!appState.pendingScoreRequest || appState.pendingScoreRequest.signature !== signature) {
            appState.pendingScoreRequest = { signature, requestId: newRequestId() };
          }
          payload.requestId = appState.pendingScoreRequest.requestId;
          await scoreEntryCreate(payload);
          appState.pendingScoreRequest = null;
          toast("已录入");
          await loadRecentScoreItems();

          await Promise.all([loadStudentsForPickers(), loadGroups()]);
          if (
            appState.scoreDraft?.scope === payload.scope &&
            (appState.scoreDraft?.targetId || 0) === (payload.targetId || 0)
          ) {
            clearScoreDraft();
          }
          render();
        } catch (e) {
          toast(String(e.message || e));
        } finally {
          button.disabled = false;
        }
      },
    });
    return button;
  }


  const recentPlus = el("div", { class: "row" });
  const recentMinus = el("div", { class: "row" });
  const allPlus = el("div", { class: "row" });
  const allMinus = el("div", { class: "row" });

  function renderItems() {
    const dimId = Number(filterDim.value || 0);
    const recentItems = (appState.recentScoreItems || []).filter((x) => (dimId ? Number(x.dimensionId) === dimId : true));
    const allItems = (appState.scoreItems || []).filter((x) => (dimId ? Number(x.dimensionId) === dimId : true));

    recentPlus.innerHTML = "";
    recentMinus.innerHTML = "";
    allPlus.innerHTML = "";
    allMinus.innerHTML = "";

    for (const it of recentItems) {
      const score = Number(it.score || 0);
      (score >= 0 ? recentPlus : recentMinus).appendChild(itemBtn(it));
    }
    for (const it of allItems) {
      const score = Number(it.score || 0);
      (score >= 0 ? allPlus : allMinus).appendChild(itemBtn(it));
    }
  }

  filterDim.addEventListener("change", () => {
    appState.scoreItemFilterDimensionId = Number(filterDim.value || 0);
    renderItems();
  });

  renderItems();

  const draftInfo = appState.scoreDraft?.targetId
    ? el("div", { class: "tag" }, [
        el("span", { text: "快捷录入：已选学生 ID" }),
        el("strong", { text: String(appState.scoreDraft.targetId) }),
        el("button", {
          class: "btn btn-small",
          text: "清除",
          onclick: () => {
            clearScoreDraft();
            render();
          },
        }),
      ])
    : null;

  const content = el("div", { class: "grid cols-2" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "录入设置" }),
      el("div", { class: "row" }, [
        el("div", { class: "field" }, [el("label", { text: "范围" }), scopeSel]),
        el("div", { class: "field" }, [el("label", { text: "对象" }), targetSel]),
        el("div", { class: "field" }, [el("label", { text: "对象搜索" }), targetKw]),
        el("div", { class: "field" }, [el("label", { text: "备注" }), remark]),
        el("div", { class: "field" }, [el("label", { text: "维度筛选" }), filterDim]),
        draftInfo,
        currentScore,
      ]),
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: "最近使用" }),
      el("div", { class: "grid", style: "gap:10px" }, [
        el("div", { class: "row" }, [el("span", { class: "pill" }, [el("span", { text: "加分项" })]), recentPlus]),
        el("div", { class: "row" }, [el("span", { class: "pill" }, [el("span", { text: "扣分项" })]), recentMinus]),
      ]),
    ]),
    el("div", { class: "card", style: "grid-column:1/-1" }, [
      el("h2", { text: "全部积分项" }),
      el("div", { class: "grid", style: "gap:10px" }, [
        el("div", { class: "row" }, [el("span", { class: "pill" }, [el("span", { text: "加分项" })]), allPlus]),
        el("div", { class: "row" }, [el("span", { class: "pill" }, [el("span", { text: "扣分项" })]), allMinus]),
      ]),
    ]),
  ]);

  return shell("积分录入", content);
}

function viewRollcall() {
  const fair = el("input", { type: "checkbox" });
  const pickCount = el("input", { type: "number", min: "1", max: "50", step: "1", placeholder: "1" });
  pickCount.value = String(appState.rollcallPickCount || 1);
  pickCount.addEventListener("change", () => {
    const v = Math.max(1, Math.min(50, Math.floor(Number(pickCount.value || 1))));
    appState.rollcallPickCount = v;
    pickCount.value = String(v);
  });
  const start = el("button", {
    class: "btn btn-amber",
    text: "开始（并点名一次）",
    onclick: async () => {
      try {
        await loadStudentsForPickers();
        await rollcallStart(fair.checked);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });
  const pick = el("button", {
    class: "btn",
    text: "再点一次",
    onclick: async () => {
      try {
        await rollcallPick();
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });
  const reset = el("button", {
    class: "btn",
    text: "重置本轮",
    onclick: async () => {
      try {
        await rollcallReset(appState.rollcall.roundId);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  const sts = appState.rollcall.students || [];
  const chosen = sts.length
    ? el("div", { class: "grid" }, [
        el("div", { class: "pill" }, [el("span", { text: `本次点到 ${sts.length} 人 · 剩余 ${appState.rollcall.remaining}` })]),
        ...sts.map((st) => {
          const jump = el("button", {
            class: "btn btn-small btn-amber",
            text: "给TA录入积分",
            onclick: async () => {
              appState.scoreDraft = { scope: "student", targetId: st.id };
              appState.route = "score";
              try {
                await Promise.all([loadStudentsForPickers(), loadGroups(), loadDimensions(), loadScoreItems(), loadRecentScoreItems()]);
              } catch (e) {
                toast(String(e.message || e));
              }
              render();
            },
          });
          return el("div", { class: "student-item" }, [
            el("div", { class: "student-name" }, [
              el("div", { class: "combo", text: `${st.studentNo} ${st.name}` }),
              el("div", { class: "meta", text: st.position || "" }),
            ]),
            el("div", { class: "row" }, [jump]),
          ]);
        }),
      ])
    : el("div", { class: "pill" }, [el("span", { text: "尚未点名" })]);

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "设置" }),
      el("div", { class: "row" }, [
        el("span", { class: "pill" }, [fair, el("span", { text: "公平模式" })]),
        el("div", { class: "field" }, [el("label", { text: "点名人数" }), pickCount]),
        start,
        pick,
        reset,
      ]),
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: "本次结果" }),
      chosen,
    ]),
  ]);

  return shell("随机点名", content);
}

function viewGroups() {
  const name = el("input", { type: "text", placeholder: "如：第一组" });
  const create = el("button", {
    class: "btn btn-amber",
    text: "新增小组",
    onclick: async () => {
      try {
        const v = (name.value || "").trim();
        if (!v) throw new Error("请输入小组名称");
        await groupCreate(v);
        name.value = "";
        toast("已创建");
        await loadGroups();
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  const list = el(
    "div",
    { class: "list" },
    (appState.groups || []).map((g) => {
      const edit = el("input", { type: "text", value: g.name });
      const save = el("button", {
        class: "btn btn-small btn-amber",
        text: "保存",
        onclick: async () => {
          try {
            const v = (edit.value || "").trim();
            if (!v) throw new Error("名称不能为空");
            await groupUpdate(g.id, v);
            toast("已保存");
            await loadGroups();
            render();
          } catch (e) {
            toast(String(e.message || e));
          }
        },
      });
      const del = el("button", {
        class: "btn btn-small btn-danger",
        text: "删除",
        onclick: async () => {
          try {
            const ok = window.confirm(`确认删除小组「${g.name}」？`);
            if (!ok) return;
            await groupDelete(g.id);
            toast("已删除");
            await loadGroups();
            render();
          } catch (e) {
            toast(String(e.message || e));
          }
        },
      });

      const avgScore = Number(g.avgScore || 0);
      const avgScoreEl = el("div", { class: avgScore >= 0 ? "score pos" : "score neg", text: String(avgScore) });
      return el("div", { class: "row entry-item" }, [
        el("div", { class: "row", style: "flex:1" }, [
          el("div", { class: "field" }, [el("label", { text: "小组名称" }), edit]),
          el("div", { class: "row" }, [el("span", { class: "pill" }, [el("span", { text: "平均分" })]), avgScoreEl]),
        ]),
        save,
        del,
      ]);
    })
  );

  const kw = el("input", { type: "text", placeholder: "搜索：姓名/学号", value: appState.groupsStudentKeyword || "" });
  kw.addEventListener("input", () => {
    appState.groupsStudentKeyword = (kw.value || "").trim();
    render();
  });

  const keyword = (appState.groupsStudentKeyword || "").trim();
  const keywordLower = keyword.toLowerCase();
  const filteredStudents = (appState.students.items || []).filter((s) => {
    if (!keywordLower) return true;
    const name = String(s.name || "");
    const no = String(s.studentNo || "");
    return name.toLowerCase().includes(keywordLower) || no.toLowerCase().includes(keywordLower);
  });

  const studentList = el(
    "div",
    { class: "list" },
    filteredStudents.map((s, idx) => {
      const groupSel = el("select");
      groupSel.appendChild(el("option", { value: "0", text: "未分组" }));
      for (const g of appState.groups || []) {
        groupSel.appendChild(el("option", { value: String(g.id), text: g.name }));
      }
      const current = String(s.groupId || 0);
      if ([...groupSel.options].some((o) => o.value === current)) groupSel.value = current;

      const save = el("button", {
        class: "btn btn-small",
        text: "保存分组",
        onclick: async () => {
          try {
            const gid = Number(groupSel.value || 0);
            await studentAssignGroup(s.id, gid);
            toast("已更新");
            await Promise.all([loadStudentsForPickers(), loadGroups()]);
            render();
          } catch (e) {
            toast(String(e.message || e));
          }
        },
      });

      const n = pad2(idx + 1);

      const combo = `${n} ${s.name}`;
      return el("div", { class: "student-item" }, [
        el("div", { class: "student-name" }, [
          el("div", { class: "combo", text: combo }),
          el("div", { class: "meta", text: `${s.studentNo} · 当前：${groupNameById(s.groupId)}` }),
        ]),
        el("div", { class: "row" }, [groupSel, save]),
      ]);
    })
  );

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "新增小组" }),
      el("div", { class: "row" }, [name, create, el("span", { class: "muted", text: "（删除小组后，该组学生将自动变为未分组）" })]),
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: "小组列表" }),
      list,
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: "学生分组" }),
      el("div", { class: "row" }, [kw]),
      studentList,
    ]),
  ]);

  return shell("小组管理", content);
}

function viewConfig() {
  const dimName = el("input", { type: "text", placeholder: "如：课堂纪律" });
  const dimCreate = el("button", {
    class: "btn btn-amber",
    text: "新增维度",
    onclick: async () => {
      try {
        const v = (dimName.value || "").trim();
        if (!v) throw new Error("请输入维度名称");
        await dimensionCreate(v);
        dimName.value = "";
        toast("已创建");
        await loadDimensions();
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });
  
  const dimList = el(
    "div",
    { class: "list" },
    (appState.dimensions || []).map((d) => {
      const edit = el("input", { type: "text", value: d.name });
      const save = el("button", {
        class: "btn btn-small btn-amber",
        text: "保存",
        onclick: async () => {
          try {
            const v = (edit.value || "").trim();
            if (!v) throw new Error("名称不能为空");
            await dimensionUpdate(d.id, v);
            toast("已保存");
            await loadDimensions();
            render();
          } catch (e) {
            toast(String(e.message || e));
          }
        },
      });
      const del = el("button", {
        class: "btn btn-small btn-danger",
        text: "删除",
        onclick: async () => {
          try {
            const ok = window.confirm(`确认删除维度「${d.name}」？`);
            if (!ok) return;
            await dimensionDelete(d.id);
            toast("已删除");
            await Promise.all([loadDimensions(), loadScoreItems()]);
            render();
          } catch (e) {
            toast(String(e.message || e));
          }
        },
      });
      const stats = el("button", {
        class: "btn btn-small",
        text: "统计",
        onclick: async () => {
          try {
            const now = new Date();
            const start = new Date(now.getFullYear(), now.getMonth(), 1);
            const params = {
              dimensionId: Number(d.id),
              startDate: appState.dimensionStats.startDate || `${start.getFullYear()}-${pad2(start.getMonth() + 1)}-01`,
              endDate: appState.dimensionStats.endDate || `${now.getFullYear()}-${pad2(now.getMonth() + 1)}-${pad2(now.getDate())}`,
            };
            await loadDimensionStats(params);
            render();
          } catch (e) {
            toast(String(e.message || e));
          }
        },
      });
      return el("div", { class: "row entry-item" }, [
        el("div", { class: "row", style: "flex:1" }, [
          el("div", { class: "field" }, [el("label", { text: "维度名称" }), edit]),
        ]),
        stats,
        save,
        del,
      ]);
    })
  );

  const selectedStatsDimension = (appState.dimensions || []).find(
    (d) => Number(d.id) === Number(appState.dimensionStats.dimensionId)
  );
  let statsCard = null;
  if (selectedStatsDimension) {
    const statsStart = el("input", { type: "date", value: appState.dimensionStats.startDate });
    const statsEnd = el("input", { type: "date", value: appState.dimensionStats.endDate });
    const statsQuery = el("button", {
      class: "btn btn-amber",
      text: "查询统计",
      onclick: async () => {
        try {
          await loadDimensionStats({
            dimensionId: Number(selectedStatsDimension.id),
            startDate: statsStart.value,
            endDate: statsEnd.value,
          });
          render();
        } catch (e) {
          toast(String(e.message || e));
        }
      },
    });
    const activeStudentIds = new Set((appState.students.items || []).map((student) => Number(student.id)));
    const items = (appState.dimensionStats.items || []).filter((item) => activeStudentIds.has(Number(item.student.id)));
    const added = items.filter((item) => Number(item.addedScore || 0) > 0);
    const deducted = items.filter((item) => Number(item.deductedScore || 0) < 0);
    const unchanged = items.filter((item) => Number(item.entryCount || 0) === 0);
    const statsList = (list, field, emptyText) =>
      el(
        "div",
        { class: "stats-list" },
        list.length
          ? list.map((item) => {
              const score = Number(item[field] || 0);
              return el("div", { class: "stats-student" }, [
                el("span", { text: `${item.student.studentNo} ${item.student.name}` }),
                field === "entryCount"
                  ? el("span", { class: "muted", text: "无记录" })
                  : el("strong", { class: score > 0 ? "score pos" : "score neg", text: `${score > 0 ? "+" : ""}${score}` }),
              ]);
            })
          : [el("div", { class: "muted", text: emptyText })]
      );
    statsCard = el("div", { class: "card" }, [
      el("div", { class: "stats-heading" }, [
        el("h2", { text: `${selectedStatsDimension.name}统计` }),
        el("div", { class: "row" }, [
          el("div", { class: "field" }, [el("label", { text: "开始日期" }), statsStart]),
          el("div", { class: "field" }, [el("label", { text: "结束日期" }), statsEnd]),
          statsQuery,
        ]),
      ]),
      el("div", { class: "stats-grid" }, [
        el("section", { class: "stats-section" }, [el("h3", { text: `有加分（${added.length}）` }), statsList(added, "addedScore", "该时段没有加分记录")]),
        el("section", { class: "stats-section" }, [el("h3", { text: `有减分（${deducted.length}）` }), statsList(deducted, "deductedScore", "该时段没有减分记录")]),
        el("section", { class: "stats-section" }, [el("h3", { text: `无加减分（${unchanged.length}）` }), statsList(unchanged, "entryCount", "全班学生均有记录")]),
      ]),
    ]);
  }

  const siDim = el("select");
  for (const d of appState.dimensions || []) {
    siDim.appendChild(el("option", { value: String(d.id), text: d.name }));
  }
  const siName = el("input", { type: "text", placeholder: "如：积极回答" });
  const siScore = el("input", { type: "number", placeholder: "分值（可正可负）" });
  siScore.value = "1";
  const siCreate = el("button", {
    class: "btn btn-amber",
    text: "新增积分项",
    onclick: async () => {
      try {
        const dimId = Number(siDim.value || 0);
        const name = (siName.value || "").trim();
        const score = Number(siScore.value);
        if (!dimId) throw new Error("请选择维度");
        if (!name) throw new Error("请输入积分项名称");
        if (Number.isNaN(score) || !Number.isFinite(score) || score === 0) throw new Error("请输入非 0 分值");
        await scoreItemCreate({ dimensionId: dimId, name, score });
        siName.value = "";
        toast("已创建");
        await loadScoreItems();
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  const filterDim = el("select");
  filterDim.appendChild(el("option", { value: "", text: "全部维度" }));
  for (const d of appState.dimensions || []) {
    filterDim.appendChild(el("option", { value: String(d.id), text: d.name }));
  }
  const scoreItemsWrap = el("div", { class: "list" });

  function renderScoreItems() {
    scoreItemsWrap.innerHTML = "";
    const filter = filterDim.value ? Number(filterDim.value) : 0;
    const items = (appState.scoreItems || []).filter((x) => (filter ? Number(x.dimensionId) === filter : true));
    if (items.length === 0) {
      scoreItemsWrap.appendChild(el("div", { class: "pill" }, [el("span", { text: "暂无积分项" })]));
      return;
    }
    for (const it of items) {
      const score = Number(it.score || 0);
      const scoreCls = score >= 0 ? "score pos" : "score neg";
      const editDim = el("select");
      for (const d of appState.dimensions || []) {
        editDim.appendChild(el("option", { value: String(d.id), text: d.name }));
      }
      if ([...editDim.options].some((o) => o.value === String(it.dimensionId))) editDim.value = String(it.dimensionId);

      const editName = el("input", { type: "text", value: it.name });
      const editScore = el("input", { type: "number", value: String(score) });
      const save = el("button", {
        class: "btn btn-small btn-amber",
        text: "保存",
        onclick: async () => {
          try {
            const dimId = Number(editDim.value || 0);
            const name = (editName.value || "").trim();
            const sc = Number(editScore.value);
            if (!dimId) throw new Error("请选择维度");
            if (!name) throw new Error("请输入积分项名称");
            if (Number.isNaN(sc) || !Number.isFinite(sc) || sc === 0) throw new Error("请输入非 0 分值");
            await scoreItemUpdate(it.id, { dimensionId: dimId, name, score: sc });
            toast("已保存");
            await loadScoreItems();
            render();
          } catch (e) {
            toast(String(e.message || e));
          }
        },
      });
      const del = el("button", {
        class: "btn btn-small btn-danger",
        text: "删除",
        onclick: async () => {
          try {
            const ok = window.confirm(`确认删除积分项「${it.name}」？`);
            if (!ok) return;
            await scoreItemDelete(it.id);
            toast("已删除");
            await Promise.all([loadScoreItems(), loadRecentScoreItems()]);
            render();
          } catch (e) {
            toast(String(e.message || e));
          }
        },
      });

      scoreItemsWrap.appendChild(
        el("div", { class: "entry-item" }, [
          el("div", { class: "row", style: "gap:8px;flex-wrap:wrap" }, [
            el("span", { class: "pill" }, [el("span", { text: "维度" })]),
            editDim,
            el("span", { class: "pill" }, [el("span", { text: "名称" })]),
            editName,
            el("span", { class: "pill" }, [el("span", { text: "分值" })]),
            editScore,
            el("div", { class: scoreCls, text: `${score >= 0 ? "+" : ""}${score}` }),
            save,
            del,
          ]),
        ])
      );
    }
  }

  filterDim.addEventListener("change", renderScoreItems);
  renderScoreItems();

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "维度" }),
      el("div", { class: "row" }, [dimName, dimCreate]),
      el("div", { class: "sep" }),
      dimList,
    ]),
    statsCard,
    el("div", { class: "card" }, [
      el("h2", { text: "新增积分项" }),
      el("div", { class: "row" }, [
        el("div", { class: "field" }, [el("label", { text: "维度" }), siDim]),
        el("div", { class: "field" }, [el("label", { text: "名称" }), siName]),
        el("div", { class: "field" }, [el("label", { text: "分值" }), siScore]),
        siCreate,
      ]),
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: "积分项列表" }),
      el("div", { class: "row" }, [el("div", { class: "field" }, [el("label", { text: "筛选" }), filterDim])]),
      scoreItemsWrap,
    ]),
  ]);

  return shell("维度与积分项", content);
}

function viewEntries() {
  const q = { ...appState.scoreEntriesQuery };

  function tokenizeKw(v) {
    const s = String(v || "").trim().toLowerCase();
    if (!s) return [];
    return s.split(/\s+/).filter(Boolean);
  }

  function matchAllTokens(haystack, tokens) {
    if (!tokens.length) return true;
    const h = String(haystack || "").toLowerCase();
    return tokens.every((t) => h.includes(t));
  }

  const stuSel = el("select");
  const stuKw = el("input", { type: "text", placeholder: "搜索（学号/姓名）" });
  stuKw.value = String(appState.scoreEntriesStudentKeyword || "");
  stuSel.appendChild(el("option", { value: "0", text: "全部学生" }));

  function refreshStudents() {
    stuSel.disabled = false;
    const prev = String(stuSel.value || "");
    stuSel.innerHTML = "";
    stuSel.appendChild(el("option", { value: "0", text: "全部学生" }));

    const kwTokens = tokenizeKw(stuKw.value);
    for (const s of appState.students.items || []) {
      const txt = `${s.studentNo || ""} ${s.name || ""}`.trim();
      if (!matchAllTokens(txt, kwTokens)) continue;
      stuSel.appendChild(el("option", { value: String(s.id), text: txt }));
    }

    if (stuSel.options.length <= 1) {
      stuSel.appendChild(el("option", { value: "", text: "无匹配学生" }));
      stuSel.value = "";
      stuSel.disabled = true;
      return;
    }

    const desired = String(q.studentId || 0);
    if (desired && [...stuSel.options].some((o) => o.value === desired)) {
      stuSel.value = desired;
      return;
    }
    if (prev && [...stuSel.options].some((o) => o.value === prev)) {
      stuSel.value = prev;
      return;
    }
    stuSel.value = "0";
  }

  refreshStudents();
  stuKw.addEventListener("input", () => {
    appState.scoreEntriesStudentKeyword = stuKw.value || "";
    refreshStudents();
  });

  const groupSel = el("select");
  groupSel.appendChild(el("option", { value: "0", text: "全部小组" }));
  for (const g of appState.groups || []) {
    groupSel.appendChild(el("option", { value: String(g.id), text: g.name }));
  }
  groupSel.value = String(q.groupId || 0);

  const sinceDays = el("input", { type: "number", min: "0", max: "36500", placeholder: "0 表示全部" });
  sinceDays.value = String(q.sinceDays ?? 30);

  const size = el("input", { type: "number", min: "1", max: "200" });
  size.value = String(q.size || 20);

  const query = el("button", {
    class: "btn btn-amber",
    text: "查询",
    onclick: async () => {
      try {
        if (stuSel.disabled) {
          throw new Error("请选择学生");
        }
        appState.scoreEntriesQuery = {
          page: 1,
          size: Math.max(1, Math.min(200, Number(size.value || 20))),
          studentId: Number(stuSel.value || 0),
          groupId: Number(groupSel.value || 0),
          sinceDays: Math.max(0, Math.min(36500, Number(sinceDays.value || 0))),
        };
        await loadScoreEntries(appState.scoreEntriesQuery);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  const prev = el("button", {
    class: "btn",
    text: "上一页",
    onclick: async () => {
      try {
        const cur = appState.scoreEntriesQuery.page || 1;
        if (cur <= 1) return;
        appState.scoreEntriesQuery = { ...appState.scoreEntriesQuery, page: cur - 1 };
        await loadScoreEntries(appState.scoreEntriesQuery);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });
  const next = el("button", {
    class: "btn",
    text: "下一页",
    onclick: async () => {
      try {
        const cur = appState.scoreEntriesQuery.page || 1;
        const sz = appState.scoreEntriesQuery.size || 20;
        const total = appState.scoreEntries.total || 0;
        if (cur * sz >= total) return;
        appState.scoreEntriesQuery = { ...appState.scoreEntriesQuery, page: cur + 1 };
        await loadScoreEntries(appState.scoreEntriesQuery);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });


  const pager = el("div", { class: "row" }, [
    el("span", { class: "pill" }, [el("span", { text: `第 ${appState.scoreEntriesQuery.page || 1} 页` })]),
    el("span", { class: "pill" }, [el("span", { text: `共 ${appState.scoreEntries.total || 0} 条` })]),
    prev,
    next,

  ]);


  const selectedStudentId = Number(stuSel.value || 0);
  const selectedStudent = selectedStudentId
    ? (appState.students.items || []).find((x) => Number(x.id) === selectedStudentId) || null
    : null;
  const selectedTotal = selectedStudent ? Number(selectedStudent.totalScore || 0) : 0;
  const currentTotalPill = selectedStudent
    ? el("span", { class: "pill" }, [el("span", { text: `当前总分（全部明细）${selectedTotal}` })])
    : null;


  const entries = el(
    "div",
    { class: "list" },
    (appState.scoreEntries.items || []).map((e) => {
      const item = scoreItemById(e.scoreItemId);
      const score = Number(e.score || 0);
      const scoreCls = score >= 0 ? "score pos" : "score neg";
      const when = e.createdAt ? new Date(Number(e.createdAt) * 1000).toLocaleString() : "";
      const itemName = e.scoreItemNameSnapshot || (item ? item.name : `积分项ID ${e.scoreItemId}`);
      const dimName = e.dimensionNameSnapshot || (item ? dimensionNameById(item.dimensionId) : dimensionNameById(e.dimensionId));
      const studentName = e.studentNameSnapshot || studentNameById(e.studentId);
      const groupName = e.groupNameSnapshot || groupNameById(e.groupId);
      const remark = (e.remark || "").trim();


      const canRevoke = true;
      const revokeBtn = canRevoke
        ? el("button", {
            class: "btn btn-small btn-danger",
            text: "撤销",
            onclick: async () => {
              try {
                const ok = window.confirm("确认撤销这条积分记录？");
                if (!ok) return;
                await scoreEntryDelete(e.id);
                toast("已撤销");
                await Promise.all([loadStudentsForPickers(), loadGroups(), loadScoreEntries(appState.scoreEntriesQuery)]);
                render();
              } catch (err) {
                toast(String(err.message || err));
              }
            },
          })
        : null;

      return el("div", { class: "entry-item" }, [
        el("div", { class: "kv" }, [
          el("div", { class: "row", style: "gap:8px;flex-wrap:wrap" }, [
            el("strong", { text: studentName }),
            el("span", { class: "pill" }, [el("span", { text: groupName })]),
            el("span", { class: "pill" }, [el("span", { text: dimName })]),
            el("span", { class: "pill" }, [el("span", { text: itemName })]),
          ]),
          el("div", { class: scoreCls, text: `${score >= 0 ? "+" : ""}${score}` }),

        ]),
        el("div", { class: "row", style: "justify-content:space-between" }, [
          el("span", { class: "muted", text: when }),
          remark ? el("span", { class: "muted", text: `备注：${remark}` }) : el("span"),
        ]),
        revokeBtn ? el("div", { class: "row", style: "justify-content:flex-end" }, [revokeBtn]) : null,
      ]);
    })
  );

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "筛选" }),
      el("div", { class: "row" }, [
        currentTotalPill,
        el("div", { class: "field" }, [el("label", { text: "学生" }), stuSel]),
        el("div", { class: "field" }, [el("label", { text: "学生搜索" }), stuKw]),
        el("div", { class: "field" }, [el("label", { text: "小组" }), groupSel]),
        el("div", { class: "field" }, [el("label", { text: "最近天数（0=全部）" }), sinceDays]),
        el("div", { class: "field" }, [el("label", { text: "每页条数" }), size]),
        query,
      ]),
      el("div", { class: "sep" }),
      pager,
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: "积分记录" }),
      entries,
    ]),
  ]);

  return shell("积分记录", content);
}

function viewRankings() {
  const typeSel = el("select", {}, [
    el("option", { value: "students", text: "学生排行" }),
    el("option", { value: "groups", text: "小组排行（平均分）" }),
  ]);
  typeSel.value = appState.rankingsType || "students";

  const now = new Date();
  const monthStart = new Date(now.getFullYear(), now.getMonth(), 1);
  const startDate = el("input", { type: "date", value: `${monthStart.getFullYear()}-${pad2(monthStart.getMonth() + 1)}-${pad2(monthStart.getDate())}` });
  const endDate = el("input", { type: "date", value: `${now.getFullYear()}-${pad2(now.getMonth() + 1)}-${pad2(now.getDate())}` });
  const total = el("input", { type: "checkbox" });
  const dim = el("select");
  dim.appendChild(el("option", { value: "", text: "全部维度" }));
  for (const d of appState.dimensions || []) {
    dim.appendChild(el("option", { value: String(d.id), text: d.name }));
  }
  const syncTotal = () => {
    const on = !!total.checked;
    startDate.disabled = on;
    endDate.disabled = on;
    dim.disabled = on;
  };
  total.addEventListener("change", () => {
    syncTotal();
  });
  typeSel.addEventListener("change", () => {
    appState.rankingsType = typeSel.value;
    syncTotal();
    render();
  });
  syncTotal();
  const topN = el("input", { type: "number", min: "0", placeholder: "前 N 名高亮（可选）" });
  const query = el("button", {
    class: "btn btn-amber",
    text: "查询",
    onclick: async () => {
      try {
        const params = {
          total: total.checked,
          startDate: startDate.value,
          endDate: endDate.value,
          dimensionId: dim.value ? Number(dim.value) : "",
          topN: topN.value ? Number(topN.value) : "",
        };
        appState.rankingsType = typeSel.value;
        if (typeSel.value === "groups") await loadGroupRankings(params);
        else await loadRankings(params);
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  const exportBtn = el("button", {
    class: "btn",
    text: "导出量化汇总及细则",
    onclick: () => {
      try {
        if (typeSel.value === "groups") {
          toast("小组排行暂不支持导出");
          return;
        }
        downloadWithAuth(
          "/rankings/students/export",
          {
            total: total.checked,
            startDate: startDate.value,
            endDate: endDate.value,
            dimensionId: dim.value ? Number(dim.value) : "",
            topN: topN.value ? Number(topN.value) : "",
          },
          total.checked ? "班级量化汇总及细则-全部历史.xlsx" : "班级量化汇总及细则.xlsx"
        );
      } catch (e) {
        toast(String(e.message || e));
      }
    }
  });

  const list = el(
    "div",
    { class: "list" },
    (typeSel.value === "groups" ? appState.groupRankings : appState.rankings || []).map((it) => {
      if (typeSel.value === "groups") {
        const g = it.group;
        const score = Number(it.score || 0);
        const scoreCls = score >= 0 ? "score pos" : "score neg";
        const name = `${pad2(it.rank)} ${g.name}`;
        return el("div", { class: "student-item", style: it.highlight ? "outline:3px solid rgba(242, 178, 75, 0.35);" : "" }, [
          el("div", { class: "student-name" }, [
            el("div", { class: "combo", text: name }),
            el("div", { class: "meta", text: "平均分" }),
          ]),
          el("div", { class: scoreCls, text: String(score) }),
        ]);
      }
      const s = it.student;
      const score = Number(it.score || 0);
      const scoreCls = score >= 0 ? "score pos" : "score neg";
      const name = `${pad2(it.rank)} ${s.name}`;
      const meta = `${s.studentNo}${s.position ? ` · ${s.position}` : ""}`;
      return el("div", { class: "student-item", style: it.highlight ? "outline:3px solid rgba(242, 178, 75, 0.35);" : "" }, [
        el("div", { class: "student-name" }, [
          el("div", { class: "combo", text: name }),
          el("div", { class: "meta", text: meta }),
        ]),
        el("div", { class: scoreCls, text: String(score) }),
      ]);
    })
  );

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "筛选" }),
      el("div", { class: "row" }, [
        el("div", { class: "field" }, [el("label", { text: "类型" }), typeSel]),
        el("div", { class: "field" }, [el("label", { text: "总分榜" }), total]),
        el("div", { class: "field" }, [el("label", { text: "开始日期" }), startDate]),
        el("div", { class: "field" }, [el("label", { text: "结束日期" }), endDate]),
        el("div", { class: "field" }, [el("label", { text: "维度" }), dim]),
        el("div", { class: "field" }, [el("label", { text: "高亮阈值" }), topN]),
        query,
        exportBtn,
      ]),
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: typeSel.value === "groups" ? "小组排行" : "学生排行" }),
      list,
    ]),
  ]);

  return shell("排行榜", content);
}

function timerTick(timerIndex) {
  const timer = appState.timers[timerIndex];
  if (!timer) return;

  stopTimerLoop(timerIndex);
  if (!timer.running) {
    syncTimerUI(timerIndex);
    return;
  }

  const now = Date.now();
  const dt = Math.max(0, now - timer.lastTick);
  timer.lastTick = now;

  if (timer.mode === "countdown") {
    timer.leftMs = Math.max(0, timer.leftMs - dt);
    if (timer.leftMs === 0) timer.running = false;
  } else {
    timer.leftMs = timer.leftMs + dt;
  }

  syncTimerUI(timerIndex);

  if (timer.running) {
    timer.timeoutId = setTimeout(() => timerTick(timerIndex), 200);
  }
}

function viewTimer() {
  function startTimer(timerIndex) {
    const timer = appState.timers[timerIndex];
    if (!timer || timer.running) return;
    timer.running = true;
    timer.lastTick = Date.now();
    syncTimerUI(timerIndex);
    timerTick(timerIndex);
  }

  function pauseTimer(timerIndex) {
    const timer = appState.timers[timerIndex];
    if (!timer || !timer.running) return;
    timer.running = false;
    stopTimerLoop(timerIndex);
    syncTimerUI(timerIndex);
  }

  function resetTimer(timerIndex) {
    const timer = appState.timers[timerIndex];
    if (!timer) return;
    timer.running = false;
    stopTimerLoop(timerIndex);
    timer.leftMs = timer.mode === "countdown" ? timer.targetMs : 0;
  }

  function buildTimerCard(timerIndex) {
    const timer = appState.timers[timerIndex];
    const uiIndex = timerIndex + 1;

    const mode = el("select", {}, [
      el("option", { value: "countdown", text: "倒计时" }),
      el("option", { value: "countup", text: "正计时" }),
    ]);
    mode.value = timer.mode;
    mode.addEventListener("change", () => {
      timer.mode = mode.value;
      timer.running = false;
      stopTimerLoop(timerIndex);
      timer.leftMs = timer.mode === "countdown" ? timer.targetMs : 0;
      render();
    });

    const customMin = el("input", {
      type: "number",
      min: "0",
      max: "240",
      step: "1",
      placeholder: "分钟",
      value: String(timer.presetMin || 5),
    });
    const customSec = el("input", {
      type: "number",
      min: "0",
      max: "59",
      step: "1",
      placeholder: "秒",
      value: String(timer.presetSec || 0),
    });
    const applyCustom = el("button", {
      class: "btn",
      text: "应用",
      onclick: () => {
        const rawM = Number(customMin.value || 0);
        const rawS = Number(customSec.value || 0);
        if (!Number.isFinite(rawM) || !Number.isFinite(rawS)) {
          toast("请输入有效时间");
          return;
        }
        const m = Math.max(0, Math.min(240, Math.floor(rawM)));
        const s = Math.max(0, Math.min(59, Math.floor(rawS)));
        if (m === 0 && s === 0) {
          toast("时间不能为 0");
          return;
        }
        timer.presetMin = m;
        timer.presetSec = s;
        timer.mode = "countdown";
        mode.value = "countdown";
        timer.targetMs = (m * 60 + s) * 1000;
        timer.leftMs = timer.targetMs;
        timer.running = false;
        stopTimerLoop(timerIndex);
        render();
      },
    });

    const presets = [1, 5, 10, 20].map((m) =>
      el("button", {
        class: "btn",
        text: `${m} 分钟`,
        onclick: () => {
          timer.presetMin = m;
          timer.presetSec = 0;
          timer.mode = "countdown";
          mode.value = "countdown";
          timer.targetMs = m * 60 * 1000;
          timer.leftMs = timer.targetMs;
          timer.running = false;
          stopTimerLoop(timerIndex);
          render();
        },
      })
    );

    const start = el("button", {
      class: "btn btn-amber",
      text: timer.running ? "暂停" : "开始",
      id: `timer-start-btn-${uiIndex}`,
      onclick: () => {
        if (timer.running) {
          timer.running = false;
          stopTimerLoop(timerIndex);
          syncTimerUI(timerIndex);
          return;
        }
        startTimer(timerIndex);
      },
    });
    const reset = el("button", {
      class: "btn",
      text: "重置",
      onclick: () => {
        resetTimer(timerIndex);
        render();
      },
    });

    return el("div", { class: "card timer" }, [
      el("h2", { text: `计时器 ${uiIndex}` }),
      el("div", { class: "row" }, [
        el("div", { class: "field" }, [el("label", { text: "模式" }), mode]),
        el("div", { class: "field" }, [el("label", { text: "自定义（分:秒）" }), el("div", { class: "row" }, [customMin, customSec])]),
        applyCustom,
        ...presets,
      ]),
      el("div", { class: "clock", id: `timer-clock-${uiIndex}`, text: fmtClock(timer.leftMs) }),
      el("div", { class: "row", style: "justify-content:center" }, [start, reset]),
    ]);
  }

  const fs = el("button", {
    class: "btn",
    text: document.fullscreenElement ? "退出全屏" : "全屏",
    onclick: async () => {
      try {
        if (!document.fullscreenElement) await document.documentElement.requestFullscreen();
        else await document.exitFullscreen();
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  const startBoth = el("button", {
    class: "btn btn-amber",
    text: "同时开始",
    onclick: () => {
      for (let i = 0; i < appState.timers.length; i++) {
        startTimer(i);
      }
    },
  });
  const pauseBoth = el("button", {
    class: "btn",
    text: "同时暂停",
    onclick: () => {
      for (let i = 0; i < appState.timers.length; i++) {
        pauseTimer(i);
      }
    },
  });
  const resetBoth = el("button", {
    class: "btn",
    text: "同时重置",
    onclick: () => {
      for (let i = 0; i < appState.timers.length; i++) {
        resetTimer(i);
      }
      render();
    },
  });

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [el("h2", { text: "页面控制" }), el("div", { class: "row" }, [startBoth, pauseBoth, resetBoth, fs])]),
    el("div", { class: "timer-dual" }, [buildTimerCard(0), buildTimerCard(1)]),
  ]);

  return shell("计时器", content);
}

async function ensureDataForRoute() {
  if (appState.route === "students") {
    await loadGroups();
    await loadStudentsList(appState.studentsQuery);
  }
  if (appState.route === "score") {
    await Promise.all([loadStudentsForPickers(), loadGroups(), loadDimensions(), loadScoreItems(), loadRecentScoreItems()]);
  }
  if (appState.route === "groups") {
    await Promise.all([loadStudentsForPickers(), loadGroups()]);
  }
  if (appState.route === "config") {
    await Promise.all([loadStudentsForPickers(), loadDimensions(), loadScoreItems()]);
  }
  if (appState.route === "entries") {
    await Promise.all([
      loadStudentsForPickers(),
      loadGroups(),
      loadDimensions(),
      loadScoreItems(),
      loadScoreEntries(appState.scoreEntriesQuery),
    ]);
  }
  if (appState.route === "rankings") {
    await Promise.all([loadDimensions(), loadRankings({}), loadGroupRankings({ total: true })]);
  }
}

function render() {
  const root = document.getElementById("app");
  root.innerHTML = "";
  if (appState.route === "login") {
    root.appendChild(viewLogin());
    return;
  }
  if (appState.route === "students") root.appendChild(viewStudents());
  else if (appState.route === "score") root.appendChild(viewScore());
  else if (appState.route === "rollcall") root.appendChild(viewRollcall());
  else if (appState.route === "groups") root.appendChild(viewGroups());
  else if (appState.route === "config") root.appendChild(viewConfig());
  else if (appState.route === "entries") root.appendChild(viewEntries());
  else if (appState.route === "rankings") root.appendChild(viewRankings());
  else if (appState.route === "timer") root.appendChild(viewTimer());
  else root.appendChild(viewStudents());
}

async function boot() {
  if (getToken()) {
    await loadBootstrap();
    try {
      await ensureDataForRoute();
    } catch {
    }
  }
  render();
}

boot();
