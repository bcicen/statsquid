import curses
from curses import panel

def run_menu(menu_items,x=0,y=0,name=None,border=True):
    """
    Display menu at x,y on a newly created window
    params:
      x(int) - x coordinate to create window
      y(int) - y coordinate to create window
      name(str) - optional title for menu
      border(bool) - display border around the menu
    """
    length = max(len(s) for s in menu_items) + 4
    height = len(menu_items)

    if name:
        height += 2
    current_opt = 0
    selected_opt = -1

    w = curses.newwin(height+2,length,y,x)
    w.keypad(1)
    w.refresh()

    while selected_opt == -1:
        #reverse and bold the selected option
        display_attr=[curses.A_NORMAL]*height
        display_attr[current_opt]=curses.A_REVERSE+curses.A_BOLD

        line = 1
        if name:
            w.addstr(line,2,name,curses.A_BOLD+curses.A_UNDERLINE)
            line += 2
        for i,v in enumerate(menu_items):
            w.addstr(line,2,str(v),display_attr[i])
            w.clrtoeol()
            line += 1

        if border:
            w.border()

        w.refresh()

        x = w.getch()
        if x == curses.KEY_DOWN:
            current_opt += 1

        elif x == curses.KEY_UP:
            current_opt -= 1

        #validation can be done by CR or space bar
        elif x == ord('\n') or x == 32 :
            selected_opt = current_opt

        #in case key pressed is a number
        elif x in range(ord('0'),ord('0')+height):
            current_opt = x - ord('0')
            selected_opt=current_opt

        if current_opt > len(menu_items) - 1:
            current_opt = len(menu_items) -1

        elif current_opt < 0:
            current_opt = 0

    return selected_opt
