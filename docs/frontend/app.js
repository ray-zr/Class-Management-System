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
    const msg = body && body.message ? body.message : typeof body === "string" ? body : "request failed";
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

const appState = {
  route: getToken() ? "students" : "login",
  me: null,
  studentEditor: { id: 0, studentNo: "", name: "", gender: "", phone: "", position: "", groupId: 0 },
  scoreDraft: {
    scope: "student",
    targetId: 0,
  },
  students: { total: 0, items: [] },
  studentsQuery: { page: 1, size: 20, keyword: "", groupId: 0 },
  groups: [],
  dimensions: [],
  scoreItems: [],
  recentScoreItems: [],
  rankings: [],
  scoreEntries: { total: 0, items: [] },
  scoreEntriesQuery: { page: 1, size: 20, studentId: 0, groupId: 0, sinceDays: 30 },
  rollcall: { roundId: "", student: null, remaining: 0 },
  timer: { mode: "countdown", running: false, targetMs: 5 * 60 * 1000, leftMs: 5 * 60 * 1000, lastTick: 0 },
};

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

async function loadRankings(params = {}) {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "" || Number.isNaN(v)) continue;
    q.set(k, String(v));
  }
  const res = await apiFetch(`/rankings/students${q.toString() ? `?${q}` : ""}`);
  appState.rankings = res.items || [];
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

async function rollcallStart(fair) {
  appState.rollcall = await apiFetch("/rollcall/start", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ fair: !!fair }),
  });
}

async function rollcallPick() {
  appState.rollcall = await apiFetch("/rollcall/pick", { method: "POST" });
}

async function rollcallReset(roundId) {
  await apiFetch("/rollcall/reset", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ roundId: roundId || "" }),
  });
  appState.rollcall = { roundId: "", student: null, remaining: 0 };
}

function viewLogin() {
  const username = el("input", { type: "text", value: "teacher" });
  const password = el("input", { type: "password", value: "teacher" });
  const btn = el("button", {
    class: "btn btn-amber",
    text: "登录",
    onclick: async () => {
      try {
        const res = await fetch(`${API}/auth/login`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ username: username.value, password: password.value }),
        });
        const body = await res.json();
        if (!res.ok) throw new Error(body.message || "login failed");
        setToken(body.accessToken);
        await loadBootstrap();
        appState.route = "students";
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  return el("div", { class: "main" }, [
    el("div", { class: "card", style: "max-width:520px;margin:10vh auto 0;" }, [
      el("h2", { text: "登录" }),
      el("div", { class: "grid" }, [
        el("div", { class: "field" }, [el("label", { text: "用户名" }), username]),
        el("div", { class: "field" }, [el("label", { text: "密码" }), password]),
        el("div", { class: "row" }, [btn]),
        el("p", { style: "margin:0;color:var(--muted);font-size:12px;" }, [
          "界面为后端配套的轻量前端。",
        ]),
      ]),
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
      await Promise.all([loadDimensions(), loadScoreItems()]);
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
        el("div", { class: "row" }, [
          el("span", { class: "pill" }, [el("span", { text: "API" }), el("a", { href: "/api/health", text: "/api" })]),
          logout,
        ]),
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

  const list = el("div", { class: "list" }, (appState.students.items || []).map((s, idx) => {
    const n = pad2(idx + 1);
    const combo = `${n} ${s.name}`;
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
        el("div", { class: "combo", text: combo }),
        el("div", {
          class: "meta",
          text: `${s.studentNo}${s.position ? ` · ${s.position}` : ""}${s.gender ? ` · ${s.gender}` : ""}${s.phone ? ` · ${s.phone}` : ""} · ${groupNameById(s.groupId)}`,
        }),
      ]),
      el("div", { class: "row" }, [
        el("div", { class: (Number(s.totalScore ?? 0) >= 0 ? "score pos" : "score neg"), text: String(s.totalScore ?? 0) }),
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
  const remark = el("input", { type: "text", placeholder: "备注（可选）" });

  function refreshTargets() {
    targetSel.innerHTML = "";
    if (scopeSel.value === "student") {
      for (const s of appState.students.items || []) {
        targetSel.appendChild(el("option", { value: String(s.id), text: `${s.studentNo} ${s.name}` }));
      }
    } else if (scopeSel.value === "group") {
      for (const g of appState.groups || []) {
        targetSel.appendChild(el("option", { value: String(g.id), text: g.name }));
      }
    } else {
      targetSel.appendChild(el("option", { value: "0", text: "全班" }));
    }

    if (appState.scoreDraft && appState.scoreDraft.scope === scopeSel.value) {
      const v = String(appState.scoreDraft.targetId || 0);
      if ([...targetSel.options].some((o) => o.value === v)) {
        targetSel.value = v;
      }
    }
  }

  scopeSel.addEventListener("change", refreshTargets);
  scopeSel.value = appState.scoreDraft?.scope || "student";
  refreshTargets();

  function itemBtn(it) {
    const score = Number(it.score || 0);
    const cls = score >= 0 ? "btn btn-green" : "btn btn-red";
    const txt = `${it.name} ${score >= 0 ? "+" : ""}${score}`;
    return el("button", {
      class: cls,
      text: txt,
      onclick: async () => {
        try {
          const payload = {
            scope: scopeSel.value,
            targetId: scopeSel.value === "class" ? 0 : Number(targetSel.value || 0),
            scoreItemId: Number(it.id),
            remark: remark.value || "",
          };
          await scoreEntryCreate(payload);
          toast("已录入");
          await loadRecentScoreItems();
          if (
            appState.scoreDraft?.scope === payload.scope &&
            (appState.scoreDraft?.targetId || 0) === (payload.targetId || 0)
          ) {
            clearScoreDraft();
          }
          render();
        } catch (e) {
          toast(String(e.message || e));
        }
      },
    });
  }

  const recent = el("div", { class: "row" }, (appState.recentScoreItems || []).map(itemBtn));
  const all = el("div", { class: "row" }, (appState.scoreItems || []).map(itemBtn));

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
        el("div", { class: "field" }, [el("label", { text: "备注" }), remark]),
        draftInfo,
      ]),
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: "最近使用" }),
      recent,
    ]),
    el("div", { class: "card", style: "grid-column:1/-1" }, [
      el("h2", { text: "全部积分项" }),
      all,
    ]),
  ]);

  return shell("积分录入", content);
}

function viewRollcall() {
  const fair = el("input", { type: "checkbox" });
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

  const st = appState.rollcall.student;
  const jump = el("button", {
    class: "btn btn-amber",
    text: "给TA录入积分",
    onclick: async () => {
      if (!st) return;
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
  const chosen = st
    ? el("div", { class: "student-item" }, [
        el("div", { class: "student-name" }, [
          el("div", { class: "combo", text: `${st.studentNo} ${st.name}` }),
          el("div", { class: "meta", text: st.position || "" }),
        ]),
        el("div", { class: "row" }, [
          el("div", { class: "pill" }, [el("span", { text: `剩余 ${appState.rollcall.remaining}` })]),
          jump,
        ]),
      ])
    : el("div", { class: "pill" }, [el("span", { text: "尚未点名" })]);

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "设置" }),
      el("div", { class: "row" }, [
        el("span", { class: "pill" }, [fair, el("span", { text: "公平模式" })]),
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

      const avg = el("div", { class: "pill" }, [el("span", { text: `平均分 ${Number(g.avgScore || 0)}` })]);
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

  const studentList = el(
    "div",
    { class: "list" },
    (appState.students.items || []).map((s, idx) => {
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
      return el("div", { class: "row entry-item" }, [
        el("div", { class: "row", style: "flex:1" }, [
          el("div", { class: "field" }, [el("label", { text: "维度名称" }), edit]),
        ]),
        save,
        del,
      ]);
    })
  );

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

  const stuSel = el("select");
  stuSel.appendChild(el("option", { value: "0", text: "全部学生" }));
  for (const s of appState.students.items || []) {
    stuSel.appendChild(el("option", { value: String(s.id), text: `${s.studentNo} ${s.name}` }));
  }
  stuSel.value = String(q.studentId || 0);

  const groupSel = el("select");
  groupSel.appendChild(el("option", { value: "0", text: "全部小组" }));
  for (const g of appState.groups || []) {
    groupSel.appendChild(el("option", { value: String(g.id), text: g.name }));
  }
  groupSel.value = String(q.groupId || 0);

  const sinceDays = el("input", { type: "number", min: "1", placeholder: "默认 30" });
  sinceDays.value = String(q.sinceDays || 30);

  const size = el("input", { type: "number", min: "1", max: "200" });
  size.value = String(q.size || 20);

  const query = el("button", {
    class: "btn btn-amber",
    text: "查询",
    onclick: async () => {
      try {
        appState.scoreEntriesQuery = {
          page: 1,
          size: Math.max(1, Math.min(200, Number(size.value || 20))),
          studentId: Number(stuSel.value || 0),
          groupId: Number(groupSel.value || 0),
          sinceDays: Math.max(1, Number(sinceDays.value || 30)),
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

  const entries = el(
    "div",
    { class: "list" },
    (appState.scoreEntries.items || []).map((e) => {
      const item = scoreItemById(e.scoreItemId);
      const score = Number(e.score || 0);
      const scoreCls = score >= 0 ? "score pos" : "score neg";
      const when = e.createdAt ? new Date(Number(e.createdAt) * 1000).toLocaleString() : "";
      const itemName = item ? item.name : `积分项ID ${e.scoreItemId}`;
      const dimName = item ? dimensionNameById(item.dimensionId) : dimensionNameById(e.dimensionId);
      const studentName = studentNameById(e.studentId);
      const groupName = groupNameById(e.groupId);
      const remark = (e.remark || "").trim();

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
      ]);
    })
  );

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "筛选" }),
      el("div", { class: "row" }, [
        el("div", { class: "field" }, [el("label", { text: "学生" }), stuSel]),
        el("div", { class: "field" }, [el("label", { text: "小组" }), groupSel]),
        el("div", { class: "field" }, [el("label", { text: "最近天数" }), sinceDays]),
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
  const month = el("input", { type: "month" });
  const total = el("input", { type: "checkbox" });
  const dim = el("select");
  dim.appendChild(el("option", { value: "", text: "全部维度" }));
  for (const d of appState.dimensions || []) {
    dim.appendChild(el("option", { value: String(d.id), text: d.name }));
  }
  const syncTotal = () => {
    const on = !!total.checked;
    month.disabled = on;
    dim.disabled = on;
  };
  total.addEventListener("change", () => {
    syncTotal();
  });
  syncTotal();
  const topN = el("input", { type: "number", min: "0", placeholder: "前 N 名高亮（可选）" });
  const query = el("button", {
    class: "btn btn-amber",
    text: "查询",
    onclick: async () => {
      try {
        await loadRankings({
          total: total.checked,
          month: month.value,
          dimensionId: dim.value ? Number(dim.value) : "",
          topN: topN.value ? Number(topN.value) : "",
        });
        render();
      } catch (e) {
        toast(String(e.message || e));
      }
    },
  });

  const exportBtn = el("button", {
    class: "btn",
    text: "导出 Excel",
    onclick: () => {
      try {
        downloadWithAuth(
          "/rankings/students/export",
          {
            total: total.checked,
            month: month.value,
            dimensionId: dim.value ? Number(dim.value) : "",
            topN: topN.value ? Number(topN.value) : "",
          },
          total.checked ? "总分积分排名汇总表.xlsx" : "月度积分排名汇总表.xlsx"
        );
      } catch (e) {
        toast(String(e.message || e));
      }
    }
  });

  const list = el("div", { class: "list" }, (appState.rankings || []).map((it) => {
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
  }));

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "筛选" }),
      el("div", { class: "row" }, [
        el("div", { class: "field" }, [el("label", { text: "总分榜" }), total]),
        el("div", { class: "field" }, [el("label", { text: "月份" }), month]),
        el("div", { class: "field" }, [el("label", { text: "维度" }), dim]),
        el("div", { class: "field" }, [el("label", { text: "高亮阈值" }), topN]),
        query,
        exportBtn,
      ]),
    ]),
    el("div", { class: "card" }, [
      el("h2", { text: "学生排行" }),
      list,
    ]),
  ]);

  return shell("排行榜", content);
}

function timerTick() {
  if (!appState.timer.running) return;
  const now = Date.now();
  const dt = now - appState.timer.lastTick;
  appState.timer.lastTick = now;
  if (appState.timer.mode === "countdown") {
    appState.timer.leftMs = Math.max(0, appState.timer.leftMs - dt);
    if (appState.timer.leftMs === 0) appState.timer.running = false;
  } else {
    appState.timer.leftMs = appState.timer.leftMs + dt;
  }
  render();
  requestAnimationFrame(timerTick);
}

function viewTimer() {
  const mode = el("select", {}, [
    el("option", { value: "countdown", text: "倒计时" }),
    el("option", { value: "countup", text: "正计时" }),
  ]);
  mode.value = appState.timer.mode;
  mode.addEventListener("change", () => {
    appState.timer.mode = mode.value;
    appState.timer.running = false;
    appState.timer.leftMs = appState.timer.mode === "countdown" ? appState.timer.targetMs : 0;
    render();
  });

  const presets = [1, 5, 10, 20].map((m) =>
    el("button", {
      class: "btn",
      text: `${m} 分钟`,
      onclick: () => {
        appState.timer.mode = "countdown";
        mode.value = "countdown";
        appState.timer.targetMs = m * 60 * 1000;
        appState.timer.leftMs = appState.timer.targetMs;
        appState.timer.running = false;
        render();
      },
    })
  );

  const start = el("button", {
    class: "btn btn-amber",
    text: appState.timer.running ? "暂停" : "开始",
    onclick: () => {
      appState.timer.running = !appState.timer.running;
      appState.timer.lastTick = Date.now();
      render();
      if (appState.timer.running) requestAnimationFrame(timerTick);
    },
  });
  const reset = el("button", {
    class: "btn",
    text: "重置",
    onclick: () => {
      appState.timer.running = false;
      appState.timer.leftMs = appState.timer.mode === "countdown" ? appState.timer.targetMs : 0;
      render();
    },
  });
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

  const content = el("div", { class: "grid" }, [
    el("div", { class: "card" }, [
      el("h2", { text: "模式与预设" }),
      el("div", { class: "row" }, [
        el("div", { class: "field" }, [el("label", { text: "模式" }), mode]),
        ...presets,
      ]),
    ]),
    el("div", { class: "card timer" }, [
      el("h2", { text: "显示" }),
      el("div", { class: "clock", text: fmtClock(appState.timer.leftMs) }),
      el("div", { class: "row", style: "justify-content:center" }, [start, reset, fs]),
    ]),
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
    await Promise.all([loadDimensions(), loadScoreItems()]);
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
    await Promise.all([loadDimensions(), loadRankings({})]);
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
