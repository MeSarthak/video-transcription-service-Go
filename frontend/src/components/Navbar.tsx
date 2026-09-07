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
  Input,
} from '@heroui/react';
import {
  Play,
  Cpu,
  Sun,
  Moon,
  LogOut,
  Plus,
  User as UserIcon,
  Search,
  LayoutGrid,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';

interface NavbarProps {
  onOpenUpload?: () => void;
  searchQuery?: string;
  onSearchChange?: (val: string) => void;
}

export const Navbar: React.FC<NavbarProps> = ({ onOpenUpload, searchQuery, onSearchChange }) => {
  const { user, logout, isAuthenticated } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const location = useLocation();

  const isDashboard = location.pathname === '/';
  const isArchitecture = location.pathname === '/architecture';

  return (
    <HeroNavbar
      maxWidth="full"
      isBordered
      className="bg-background/95 dark:bg-yt-dark/95 backdrop-blur-md sticky top-0 z-50 border-b border-default-200 dark:border-yt-border px-2 sm:px-6 h-16"
    >
      {/* Brand / Logo */}
      <NavbarBrand className="gap-3 max-w-fit">
        <Link to="/" className="flex items-center gap-2 group">
          <div className="w-9 h-7 rounded-lg bg-yt-red text-white flex items-center justify-center shadow-md shadow-red-600/30 group-hover:scale-105 transition-transform">
            <Play className="w-4 h-4 fill-white ml-0.5" />
          </div>
          <div className="flex items-baseline gap-1">
            <span className="font-extrabold text-xl tracking-tight text-foreground font-sans">
              Transcribe<span className="text-yt-red">X</span>
            </span>
            <span className="text-[10px] text-default-400 font-bold tracking-wider uppercase font-mono px-1 rounded bg-default-100 dark:bg-yt-surface">
              STUDIO
            </span>
          </div>
        </Link>
      </NavbarBrand>

      {/* Center Search Pill (YouTube style) */}
      <NavbarContent className="hidden md:flex flex-1 max-w-xl mx-4" justify="center">
        <div className="w-full">
          <Input
            size="sm"
            radius="full"
            placeholder="Search your transcribed videos..."
            value={searchQuery || ''}
            onValueChange={onSearchChange}
            startContent={<Search className="w-4 h-4 text-default-400 ml-1" />}
            classNames={{
              inputWrapper:
                'bg-default-100/70 dark:bg-yt-surface border border-default-200 dark:border-yt-border hover:border-default-400 dark:hover:border-default-600 focus-within:!border-yt-red transition-all',
              input: 'text-xs text-foreground',
            }}
            isClearable
          />
        </div>
      </NavbarContent>

      {/* Right Controls */}
      <NavbarContent justify="end" className="gap-2 sm:gap-3">
        {/* Navigation Tabs */}
        <NavbarItem className="hidden sm:flex">
          <Button
            as={Link}
            to="/"
            variant="light"
            size="sm"
            radius="full"
            startContent={<LayoutGrid className="w-4 h-4" />}
            className={`text-xs font-semibold ${
              isDashboard ? 'bg-default-200/60 dark:bg-yt-surface text-foreground font-bold' : 'text-default-500 hover:text-foreground'
            }`}
          >
            Videos
          </Button>
        </NavbarItem>

        <NavbarItem>
          <Button
            as={Link}
            to="/architecture"
            variant="light"
            size="sm"
            radius="full"
            startContent={<Cpu className="w-4 h-4 text-yt-red" />}
            className={`text-xs font-semibold ${
              isArchitecture ? 'bg-default-200/60 dark:bg-yt-surface text-foreground font-bold' : 'text-default-500 hover:text-foreground'
            }`}
          >
            Architecture
          </Button>
        </NavbarItem>

        {/* YouTube "+ Create" Upload Pill Button */}
        {isAuthenticated && onOpenUpload && (
          <NavbarItem>
            <Button
              size="sm"
              radius="full"
              startContent={<Plus className="w-4 h-4 text-white stroke-[3]" />}
              onClick={onOpenUpload}
              className="bg-yt-red hover:bg-red-700 text-white font-semibold text-xs shadow-sm shadow-red-600/30 px-3.5"
            >
              Create
            </Button>
          </NavbarItem>
        )}

        {/* Theme Toggle */}
        <NavbarItem>
          <Button
            isIconOnly
            variant="light"
            size="sm"
            radius="full"
            onClick={toggleTheme}
            aria-label="Toggle theme"
            className="text-default-500 hover:text-foreground"
          >
            {theme === 'dark' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
          </Button>
        </NavbarItem>

        {/* User Account / Sign In */}
        {isAuthenticated ? (
          <NavbarItem>
            <Dropdown placement="bottom-end">
              <DropdownTrigger>
                <Avatar
                  as="button"
                  className="transition-transform ring-2 ring-yt-red/40 w-8 h-8 cursor-pointer text-xs font-bold"
                  color="danger"
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
              color="danger"
              variant="flat"
              size="sm"
              radius="full"
              startContent={<UserIcon className="w-4 h-4" />}
              className="font-semibold text-xs"
            >
              Sign In
            </Button>
          </NavbarItem>
        )}
      </NavbarContent>
    </HeroNavbar>
  );
};
