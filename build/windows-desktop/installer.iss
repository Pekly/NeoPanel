; Compile with Inno Setup 6 via build-windows-desktop.ps1 -Installer.
#ifndef SourceDir
  #error SourceDir must point to the built desktop package.
#endif
#ifndef TargetArch
  #define TargetArch "amd64"
#endif
[Setup]
AppId={{258DA21F-66FD-4ED7-B575-1EBD072F49CD}
AppName=NeoPanel Desktop
AppVersion=1.0.0-windows-beta.1
AppPublisher=NeoPanel
DefaultDirName={localappdata}\Programs\NeoPanel Desktop
DefaultGroupName=NeoPanel Desktop
PrivilegesRequired=lowest
OutputBaseFilename=NeoPanel-Desktop-Setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
UninstallDisplayIcon={app}\NeoPanel-Desktop.exe
CloseApplications=no
RestartApplications=no
#if TargetArch == "arm64"
ArchitecturesAllowed=arm64
ArchitecturesInstallIn64BitMode=arm64
#else
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
#endif

[Tasks]
Name: "desktopicon"; Description: "Create a desktop shortcut"; Flags: unchecked

[Files]
Source: "{#SourceDir}\NeoPanel-Desktop.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#SourceDir}\README-Desktop.md"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\NeoPanel Desktop"; Filename: "{app}\NeoPanel-Desktop.exe"
Name: "{group}\Uninstall NeoPanel Desktop"; Filename: "{uninstallexe}"
Name: "{userdesktop}\NeoPanel Desktop"; Filename: "{app}\NeoPanel-Desktop.exe"; Tasks: desktopicon

[Run]
Filename: "{app}\NeoPanel-Desktop.exe"; Description: "Launch NeoPanel Desktop"; Flags: nowait postinstall skipifsilent

; Intentionally no UninstallDelete entries: databases, worlds, config and WebView profile stay intact.
