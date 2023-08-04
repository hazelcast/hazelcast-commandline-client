#define MyAppName "Hazelcast CLC"
#define MyAppVersion "v5.3.2"
#define MyAppPublisher "Hazelcast, Inc."
#define MyAppURL "https://www.hazelcast.com/"
#define MyAppExeName "clc.exe"

#ifndef SourceDir
#define SourceDir "NOT-SPECIFIED"
#endif

[Setup]
; NOTE: The value of AppId uniquely identifies this application. Do not use the same AppId value in installers for other applications.
AppId={{A68F0128-173F-4144-B294-D7A76299199E}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
;AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
ChangesEnvironment=true
DisableProgramGroupPage=yes
;DefaultGroupName={#MyAppName}
LicenseFile={#SourceDir}\LICENSE
InfoBeforeFile={#SourceDir}\extras\windows\installer\pre.txt
InfoAfterFile={#SourceDir}\extras\windows\installer\post.txt
; Remove the following line to run in administrative install mode (install for all users.)
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
OutputDir={#SourceDir}\extras\windows\installer_output
OutputBaseFilename=hazelcast-clc-setup
SetupIconFile={#SourceDir}\extras\windows\installer\hazelcast_64x64.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "{#SourceDir}\build\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#SourceDir}\extras\windows\installer\README.txt"; DestDir: "{app}"; Flags: ignoreversion
; NOTE: Don't use "Flags: ignoreversion" on any shared system files

[Registry]
; update the path for all users
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; \
    ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; \
    Check: CanAddPath(true, '{app}');
; update the path for the current user
Root: HKCU; Subkey: "Environment"; \
    ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; \
    Check: CanAddPath(false, '{app}');

[Icons]
; Start Menu
Name: "{autoprograms}\{#MyAppName}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{autoprograms}\{#MyAppName}\README.txt"; Filename: "{app}\README.txt"
Name: "{autoprograms}\{#MyAppName}\Documentation"; Filename: "https://docs.hazelcast.com/hazelcast/latest-dev/clients/clc"
Name: "{autoprograms}\{#MyAppName}\Get Started with Hazelcast"; Filename: "https://docs.hazelcast.com/hazelcast/latest/getting-started/get-started-cli"
Name: "{autoprograms}\{#MyAppName}\Survey"; Filename: "https://forms.gle/rPFywdQjvib1QCe49"
; Desktop
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

; Disabled the option to automatically launch CLC for now...
;[Run]
;Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[Code]

function CanAddPath(ForAll: boolean; Param: string): boolean;
var
  OrigPath: string;
  Root: integer;
  Key: string;
begin
  { check that ForAll is being used with admin mode }
  if IsAdminInstallMode xor ForAll then begin
    Result := False;
    exit;
  end;
  if ForAll then begin
    Root := HKEY_LOCAL_MACHINE;
    Key := 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';
  end
  else begin
    Root := HKEY_CURRENT_USER;
    Key := 'Environment';
  end;    
  Param := ExpandConstant(Param);
  if not RegQueryStringValue(Root, Key, 'Path', OrigPath)
  then begin
    Result := True;
    exit;
  end;
  { look for the path with leading and trailing semicolon }
  { Pos() returns 0 if not found }
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;
