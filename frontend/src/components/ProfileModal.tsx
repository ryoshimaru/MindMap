import { useEffect, useState, type FormEvent } from "react";
import { ApiClientError } from "../api/client";
import { profileApi } from "../api/endpoints";
import type { UserProfile, UserProfileInput } from "../api/types";
import { getErrorMessage } from "../utils/errors";

type ProfileForm = {
  age: string;
  occupation: string;
  freeHoursPerWeek: string;
  availableBudget: string;
  constraints: string;
};

const emptyProfile: ProfileForm = { age: "", occupation: "", freeHoursPerWeek: "", availableBudget: "", constraints: "" };

function toProfileForm(profile: UserProfile): ProfileForm {
  return {
    age: String(profile.age),
    occupation: profile.occupation,
    freeHoursPerWeek: String(profile.freeHoursPerWeek),
    availableBudget: String(profile.availableBudget),
    constraints: profile.constraints
  };
}

export function ProfileModal({ required = false, onReady, onClose }: { required?: boolean; onReady?: () => void; onClose?: () => void }) {
  const [form, setForm] = useState(emptyProfile);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    profileApi.get().then((profile) => setForm(toProfileForm(profile))).catch((reason) => {
      if (!(reason instanceof ApiClientError && reason.status === 404)) setError(getErrorMessage(reason));
    }).finally(() => setLoading(false));
  }, []);

  async function submit(event: FormEvent) {
    event.preventDefault(); setSaving(true); setError("");
    const payload: UserProfileInput = {
      age: Number(form.age),
      occupation: form.occupation,
      freeHoursPerWeek: Number(form.freeHoursPerWeek.replace(",", ".")),
      availableBudget: Number(form.availableBudget.replace(",", ".")),
      constraints: form.constraints
    };
    try { await profileApi.save(payload); onReady?.(); onClose?.(); }
    catch (reason) { setError(getErrorMessage(reason)); }
    finally { setSaving(false); }
  }

  return <div className="modal-backdrop" role="dialog" aria-modal="true"><section className="modal-panel">
    <div className="modal-title"><div><span className="eyebrow">Профиль</span><h2>Расскажите немного о себе</h2></div>{!required && onClose ? <button className="icon-button" onClick={onClose}>×</button> : null}</div>
    <p>Эти данные помогают AI не повторять одни и те же вопросы для каждой цели.</p>
    {loading ? <p>Загрузка...</p> : <form className="form-grid" onSubmit={submit}>
      <label>Возраст<input type="text" inputMode="numeric" pattern="[0-9]+" value={form.age} onChange={(e) => setForm({...form, age: e.target.value})} required /></label>
      <label>Род занятий<input value={form.occupation} onChange={(e) => setForm({...form, occupation: e.target.value})} placeholder="Например, студент или разработчик" required /></label>
      <label>Свободное время в неделю, ч.<input type="text" inputMode="decimal" pattern="[0-9]+([.,][0-9]+)?" value={form.freeHoursPerWeek} onChange={(e) => setForm({...form, freeHoursPerWeek: e.target.value})} required /></label>
      <label>Доступный бюджет, руб.<input type="text" inputMode="decimal" pattern="[0-9]+([.,][0-9]+)?" value={form.availableBudget} onChange={(e) => setForm({...form, availableBudget: e.target.value})} required /></label>
      <label>Ограничения<textarea value={form.constraints} onChange={(e) => setForm({...form, constraints: e.target.value})} placeholder="Здоровье, график или другие важные условия" /></label>
      {error ? <div className="inline-error">{error}</div> : null}
      <button className="button button--primary" disabled={saving}>{saving ? "Сохранение..." : "Сохранить профиль"}</button>
    </form>}
  </section></div>;
}
