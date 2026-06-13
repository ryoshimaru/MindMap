import { useState } from "react";
import { NavLink } from "react-router-dom";
import { useAuth } from "../auth/useAuth";
import { AISettingsModal } from "./AISettingsModal";
import { ProfileModal } from "./ProfileModal";

export function Header() {
  const { user, logout } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);
  const [aiOpen, setAIOpen] = useState(false);

  return <>
    <header className="topbar">
      <NavLink to="/dashboard" className="wordmark">mindmap</NavLink>
      <nav className="topbar__nav" aria-label="Основная навигация">
        <NavLink to="/dashboard">Создать цель</NavLink>
        <NavLink to="/goals">Мои цели</NavLink>
      </nav>
      <div className="profile-menu">
        <button className="avatar-button" onClick={() => setMenuOpen((value) => !value)} aria-label="Открыть профиль">{user?.name?.slice(0, 1).toUpperCase() || "П"}</button>
        {menuOpen ? <div className="profile-menu__popover">
          <strong>{user?.name}</strong><span>{user?.email}</span>
          <button onClick={() => { setProfileOpen(true); setMenuOpen(false); }}>Личные данные</button>
          <button onClick={() => { setAIOpen(true); setMenuOpen(false); }}>AI-провайдер</button>
          <button onClick={logout}>Выйти</button>
        </div> : null}
      </div>
    </header>
    {profileOpen ? <ProfileModal onClose={() => setProfileOpen(false)} /> : null}
    {aiOpen ? <AISettingsModal canClose onClose={() => setAIOpen(false)} /> : null}
  </>;
}
