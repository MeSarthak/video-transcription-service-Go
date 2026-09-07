import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import {
  Navbar as HeroNavbar,
  NavbarBrand,
  NavbarContent,
  NavbarItem,
  Button,
  Dropdown,
  DropdownTrigger,
  DropdownMenu,
  DropdownItem,
  Avatar,
} from '@heroui/react';
import {
  Video as VideoIcon,
  Cpu,
  Sun,
  Moon,
  LogOut,
  Upload,
  User as UserIcon,
  Code2,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';

interface NavbarProps {
  onOpenUpload?: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ onOpenUpload }) => {
  const { user, logout, isAuthenticated } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const location = useLocation();

  const isDashboard = location.pathname === '/';
  const isArchitecture = location.pathname === '/architecture';

  return (
    <HeroNavbar
      maxWidth="xl"
      isBordered
      className="bg-background/70 backdrop-blur-md sticky top-0 z-50 shadow-sm"
    >
      <NavbarBrand>
        <Link to="/" className="flex items-center gap-2.5 group">
          <div className="p-2 rounded-xl bg-gradient-to-tr from-primary-600 to-indigo-500 text-white shadow-md shadow-primary-500/20 group-hover:scale-105 transition-transform">
            <VideoIcon className="w-5 h-5" />
          </div>
          <div className="flex flex-col">
            <span className="font-bold text-lg tracking-tight bg-gradient-to-r from-primary-500 to-indigo-400 bg-clip-text text-transparent">
              TranscribeX
            </span>
            <span className="text-[10px] text-default-400 uppercase tracking-widest font-semibold">
              Distributed Engine
            </span>
          </div>
        </Link>
      </NavbarBrand>

      <NavbarContent className="hidden sm:flex gap-4" justify="center">
        <NavbarItem isActive={isDashboard}>
          <Link
            to="/"
            className={`text-sm font-medium px-3 py-1.5 rounded-lg transition-colors ${
              isDashboard
                ? 'text-primary font-semibold bg-primary/10'
                : 'text-default-600 hover:text-foreground'
            }`}
          >
            Dashboard
          </Link>
        </NavbarItem>
        <NavbarItem isActive={isArchitecture}>
          <Link
            to="/architecture"
            className={`text-sm font-medium px-3 py-1.5 rounded-lg transition-colors flex items-center gap-1.5 ${
              isArchitecture
                ? 'text-primary font-semibold bg-primary/10'
                : 'text-default-600 hover:text-foreground'
            }`}
          >
            <Cpu className="w-4 h-4 text-primary" />
            <span>Architecture & Design</span>
          </Link>
        </NavbarItem>
      </NavbarContent>

      <NavbarContent justify="end" className="gap-2">
        <NavbarItem>
          <Button
            isIconOnly
            variant="light"
            size="sm"
            onClick={toggleTheme}
            aria-label="Toggle theme"
            className="text-default-500"
          >
            {theme === 'dark' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
          </Button>
        </NavbarItem>

        <NavbarItem>
          <Button
            isIconOnly
            as="a"
            href="https://github.com"
            target="_blank"
            variant="light"
            size="sm"
            aria-label="GitHub"
            className="text-default-500"
          >
            <Code2 className="w-4 h-4" />
          </Button>
        </NavbarItem>

        {isAuthenticated && onOpenUpload && (
          <NavbarItem>
            <Button
              color="primary"
              size="sm"
              startContent={<Upload className="w-4 h-4" />}
              onClick={onOpenUpload}
              className="font-medium shadow-md shadow-primary/20"
            >
              Upload Video
            </Button>
          </NavbarItem>
        )}

        {isAuthenticated ? (
          <NavbarItem>
            <Dropdown placement="bottom-end">
              <DropdownTrigger>
                <Avatar
                  as="button"
                  className="transition-transform ring-2 ring-primary/30 w-8 h-8 cursor-pointer"
                  color="primary"
                  name={user?.email?.slice(0, 2).toUpperCase()}
                  size="sm"
                />
              </DropdownTrigger>
              <DropdownMenu aria-label="Profile Actions" variant="flat">
                <DropdownItem key="profile" className="h-14 gap-2" textValue="Signed in as">
                  <p className="font-semibold text-xs text-default-400">Signed in as</p>
                  <p className="font-bold text-sm truncate">{user?.email}</p>
                </DropdownItem>
                <DropdownItem
                  key="logout"
                  color="danger"
                  className="text-danger"
                  startContent={<LogOut className="w-4 h-4" />}
                  onPress={logout}
                  textValue="Log Out"
                >
                  Log Out
                </DropdownItem>
              </DropdownMenu>
            </Dropdown>
          </NavbarItem>
        ) : (
          <NavbarItem>
            <Button
              as={Link}
              to="/login"
              color="primary"
              variant="flat"
              size="sm"
              startContent={<UserIcon className="w-4 h-4" />}
            >
              Sign In
            </Button>
          </NavbarItem>
        )}
      </NavbarContent>
    </HeroNavbar>
  );
};
